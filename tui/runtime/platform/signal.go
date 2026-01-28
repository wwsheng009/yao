package platform

import (
	"os"
	"os/signal"
)

// SignalHandler 信号处理抽象 (V3: 从 Terminal 拆分)
type SignalHandler interface {
	// 注册信号处理
	Handle(signals []os.Signal, handler func(os.Signal))

	// 启动监听
	Start() error

	// 停止监听
	Stop() error
}

// DefaultSignalHandler 默认信号处理实现
type DefaultSignalHandler struct {
	signals []os.Signal
	handler func(os.Signal)
	stop    chan struct{}
}

// NewDefaultSignalHandler 创建默认信号处理器
func NewDefaultSignalHandler() *DefaultSignalHandler {
	return &DefaultSignalHandler{
		stop: make(chan struct{}),
	}
}

// Handle 注册信号处理
func (s *DefaultSignalHandler) Handle(signals []os.Signal, handler func(os.Signal)) {
	s.signals = signals
	s.handler = handler
}

// Start 启动监听
func (s *DefaultSignalHandler) Start() error {
	// 创建信号通道
	sigChan := make(chan os.Signal, len(s.signals))
	signal.Notify(sigChan, s.signals...)

	// 启动监听协程
	go func() {
		for {
			select {
			case sig := <-sigChan:
				if s.handler != nil {
					s.handler(sig)
				}
			case <-s.stop:
				return
			}
		}
	}()

	return nil
}

// Stop 停止监听
func (s *DefaultSignalHandler) Stop() error {
	close(s.stop)
	return nil
}
