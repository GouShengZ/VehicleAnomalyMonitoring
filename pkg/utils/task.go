package utils

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type TaskManager struct {
	tasks    map[string]*Task
	ctx      context.Context
	cancel   context.CancelFunc
	taskLock sync.RWMutex
	wg       sync.WaitGroup
}

// Task 表示一个长期运行的任务
type Task struct {
	ID          string
	Name        string
	Status      string
	LastError   error
	StopChan    chan struct{}
	ProcessFunc func(context.Context) error
	Interval    time.Duration
	StartDelay  time.Duration
	LastRunTime time.Time
	stopOnce    sync.Once // 新增同步控制
}

// NewTaskManager 创建一个新的任务管理器
func NewTaskManager() *TaskManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskManager{
		tasks:  make(map[string]*Task),
		ctx:    ctx,
		cancel: cancel,
	}
}

// AddTask 添加一个新任务到管理器，增加了时间间隔参数
func (tm *TaskManager) AddTask(id string, name string, interval time.Duration, processFunc func(context.Context) error) error {
	tm.taskLock.Lock()
	defer tm.taskLock.Unlock()

	if _, exists := tm.tasks[id]; exists {
		return fmt.Errorf("任务 %s 已存在", id)
	}

	task := &Task{
		ID:          id,
		Name:        name,
		Status:      "ready",
		StopChan:    make(chan struct{}),
		ProcessFunc: processFunc,
		Interval:    interval,
		StartDelay:  0, // 默认不延迟
	}

	tm.tasks[id] = task
	return nil
}

// AddTaskWithDelay 添加一个新任务到管理器，支持设置首次执行延迟
func (tm *TaskManager) AddTaskWithDelay(id string, name string, interval time.Duration, startDelay time.Duration, processFunc func(context.Context) error) error {
	tm.taskLock.Lock()
	defer tm.taskLock.Unlock()

	if _, exists := tm.tasks[id]; exists {
		return fmt.Errorf("任务 %s 已存在", id)
	}

	task := &Task{
		ID:          id,
		Name:        name,
		Status:      "ready",
		StopChan:    make(chan struct{}),
		ProcessFunc: processFunc,
		Interval:    interval,
		StartDelay:  startDelay,
	}

	tm.tasks[id] = task
	return nil
}

// StartAllTasks 启动所有已注册的任务
// 返回已成功启动的任务数量和可能发生的错误
func (tm *TaskManager) StartAllTasks() (int, error) {
	tm.taskLock.RLock()
	defer tm.taskLock.RUnlock()

	successCount := 0
	var lastError error

	// 遍历所有任务并尝试启动
	for id := range tm.tasks {
		if err := tm.StartTask(id); err != nil {
			lastError = fmt.Errorf("启动任务 %s 失败: %v", id, err)
			log.Printf("启动任务 %s 失败: %v", id, err)
			continue
		}
		successCount++
	}

	// 如果有任务启动失败，返回最后一个错误
	if successCount < len(tm.tasks) {
		return successCount, lastError
	}

	return successCount, nil
}

// StartTask 启动指定的任务
func (tm *TaskManager) StartTask(id string) error {
	tm.taskLock.RLock()
	task, exists := tm.tasks[id]
	tm.taskLock.RUnlock()

	if !exists {
		return fmt.Errorf("任务 %s 不存在", id)
	}

	tm.wg.Add(1)
	go func() {
		defer tm.wg.Done()

		tm.taskLock.Lock()
		task.Status = "running"
		tm.taskLock.Unlock()

		taskCtx, cancel := context.WithCancel(tm.ctx)
		defer cancel()

		// 处理首次执行延迟
		if task.StartDelay > 0 {
			select {
			case <-taskCtx.Done():
				task.Status = "stopped"
				return
			case <-time.After(task.StartDelay):
				// 继续执行
			}
		}

		// 创建定时器
		ticker := time.NewTicker(task.Interval)
		defer ticker.Stop()

		// 首次立即执行一次
		if err := task.ProcessFunc(taskCtx); err != nil {
			task.LastError = err
			task.Status = "error"
			log.Printf("任务 %s 执行出错: %v", task.ID, err)
		}
		task.LastRunTime = time.Now()

		// 定时执行任务
		for {
			select {
			case <-taskCtx.Done():
				task.Status = "stopped"
				return
			case <-task.StopChan:
				task.Status = "stopped"
				return
			case <-ticker.C:
				if err := task.ProcessFunc(taskCtx); err != nil {
					task.LastError = err
					task.Status = "error"
					log.Printf("任务 %s 执行出错: %v", task.ID, err)
				}
				task.LastRunTime = time.Now()
			}
		}
	}()

	return nil
}

// StopTask 停止指定的任务
func (tm *TaskManager) StopTask(id string) error {
	tm.taskLock.Lock()
	defer tm.taskLock.Unlock()

	task, exists := tm.tasks[id]
	if !exists {
		return fmt.Errorf("任务 %s 不存在", id)
	}

	task.stopOnce.Do(func() {
		close(task.StopChan)
		delete(tm.tasks, id) // 移除已停止的任务
	})
	return nil
}

// StopAll 停止所有任务
func (tm *TaskManager) StopAll() {
	tm.cancel()
	tm.wg.Wait()
}
