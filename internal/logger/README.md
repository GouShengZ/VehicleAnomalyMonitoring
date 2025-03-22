# 日志使用最佳实践

本项目提供了两种日志使用方式，以避免在每个需要日志的地方都调用`configs.GetLogger()`：

## 1. 全局Logger方式

在`pkg/common/logger.go`中定义了全局Logger变量，所有文件都可以导入并使用这个变量。

### 使用方法

```go
import (
    "github.com/zhangyuchen/AutoDataHub-monitor/pkg/common"
    "go.uber.org/zap"
)

func SomeFunction() {
    common.Logger.Info("这是一条日志", zap.String("key", "value"))
}
```

### 优点

- 简单易用，只需导入一次
- 保证整个应用使用同一个Logger实例
- 便于全局配置和管理

### 缺点

- 全局变量可能导致测试困难
- 不同模块无法使用不同的日志配置

## 2. 文件级别Logger方式

在每个文件的开头定义一个局部Logger变量，并在init函数中初始化。

### 使用方法

```go
package mypackage

import (
    "github.com/zhangyuchen/AutoDataHub-monitor/configs"
    "go.uber.org/zap"
)

// 在文件级别定义logger变量
var logger *zap.Logger

// 在init函数中初始化logger
func init() {
    logger = configs.GetLogger()
}

func SomeFunction() {
    logger.Info("这是一条日志")
}
```

### 优点

- 每个文件可以有自己的logger变量
- 便于在文件内使用，代码更简洁
- 便于未来扩展（如为不同文件配置不同的logger）

### 缺点

- 每个文件都需要初始化，有一定的重复代码
- 可能导致多个Logger实例，增加内存使用

## 3. 结构体内Logger方式

对于结构体，可以在结构体中添加logger字段，并在初始化时设置。

### 使用方法

```go
type MyService struct {
    // 其他字段
    logger *zap.Logger
}

func NewMyService() *MyService {
    return &MyService{
        // 初始化其他字段
        logger: common.Logger, // 或 configs.GetLogger()
    }
}

func (s *MyService) DoSomething() {
    s.logger.Info("执行操作")
}
```

### 优点

- 结构体方法可以直接使用内部logger
- 便于测试时注入模拟logger
- 代码更加面向对象

## 推荐使用方式

1. 对于服务和大型组件，推荐使用**结构体内Logger方式**
2. 对于工具函数和小型包，推荐使用**全局Logger方式**
3. 只有在特殊需求下才考虑使用**文件级别Logger方式**