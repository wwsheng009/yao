package component

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTextComponentMeasure(t *testing.T) {
	props := TextProps{
		Content:         "Hello World",
		Width:           0, // 不指定固定宽度
	}
	
	textWrapper := NewTextComponentWrapper(props, "test-text")
	
	// 测试当内容宽度小于最大约束时
	width, height := textWrapper.Measure(50, 10) // 最大宽度50，最大高度10
	assert.Equal(t, 11, width)  // "Hello World" 的长度是11
	assert.Equal(t, 1, height)  // 单行内容高度为1
	
	// 测试当内容宽度超过最大约束时
	width, height = textWrapper.Measure(8, 10) // 最大宽度8，应被限制
	assert.Equal(t, 8, width)   // 应被限制在8
	assert.Equal(t, 1, height)  // 高度仍为1
}

func TestTextComponentMeasureWithMultiline(t *testing.T) {
	props := TextProps{
		Content:         "Line 1\nLine 2\nLine 3",
		Width:           0,
	}
	
	textWrapper := NewTextComponentWrapper(props, "test-text-multiline")
	
	// 测试多行内容
	width, height := textWrapper.Measure(20, 10)
	assert.Equal(t, 7, width)   // 最长行 "Line 1", "Line 2", "Line 3" 都是6个字符，加1为7
	assert.Equal(t, 3, height)  // 3行内容
}

func TestTextComponentMeasureWithFixedWidth(t *testing.T) {
	props := TextProps{
		Content:         "Hello World",
		Width:           15, // 指定固定宽度
	}
	
	textWrapper := NewTextComponentWrapper(props, "test-text-fixed-width")
	
	// 测试当指定了固定宽度时
	width, height := textWrapper.Measure(50, 10)
	assert.Equal(t, 15, width)  // 应使用指定的固定宽度
	assert.Equal(t, 1, height)  // 单行内容高度为1
}

func TestTextComponentMeasureWithFixedHeight(t *testing.T) {
	props := TextProps{
		Content:         "Hello",
		Width:           10,
	}
	
	textWrapper := NewTextComponentWrapper(props, "test-text-fixed-height")
	// 直接设置model的高度
	textWrapper.model.Height = 5
	
	// 测试当设置了固定高度时
	width, height := textWrapper.Measure(20, 10)
	assert.Equal(t, 10, width)  // 应使用内容宽度（5）和props宽度（10）中的较大值
	assert.Equal(t, 5, height)  // 应使用固定的model高度
}