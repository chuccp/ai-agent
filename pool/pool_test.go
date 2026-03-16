package pool

import (
	"sync/atomic"
	"testing"
	"time"
)

// TestConcurrencyLimit 测试并发限制是否有效
// 注意：caller_runs 策略下，容量用完时任务在调用者线程执行
// 所以最大并发数可能略超设定值（goroutine数 + 调用者线程）
func TestConcurrencyLimit(t *testing.T) {
	pool := NewGOPool(2)

	var maxConcurrent int32
	var currentConcurrent int32
	var taskCount int32

	for i := 0; i < 10; i++ {
		task := func() {
			cur := atomic.AddInt32(&currentConcurrent, 1)
			for {
				max := atomic.LoadInt32(&maxConcurrent)
				if cur <= max || atomic.CompareAndSwapInt32(&maxConcurrent, max, cur) {
					break
				}
			}
			time.Sleep(50 * time.Millisecond)
			atomic.AddInt32(&taskCount, 1)
			atomic.AddInt32(&currentConcurrent, -1)
		}
		pool.pool.Submit(task)
	}

	time.Sleep(1 * time.Second)

	count := atomic.LoadInt32(&taskCount)
	max := atomic.LoadInt32(&maxConcurrent)

	t.Logf("完成任务数: %d, 最大并发数: %d", count, max)

	if count != 10 {
		t.Errorf("任务数量错误: 期望 10, 实际 %d", count)
	}
	// caller_runs 策略：并发数 = goroutine数 + 可能的调用者线程
	// 2 goroutines + 1 caller = 最多 3
	if max > 3 {
		t.Errorf("并发限制失效: 期望最大并发 <= 3, 实际 %d", max)
	}
}

// TestNestedGoroutines 测试嵌套调用场景
// 注意：阻塞等待方案下，worker 数量必须 >= 嵌套深度，否则会死锁
func TestNestedGoroutines(t *testing.T) {
	// 嵌套深度为 4，需要至少 4 个 worker
	pool := NewGOPool(1)

	var taskCount int32

	err := pool.WaitGO(func() error {
		atomic.AddInt32(&taskCount, 1)
		t.Log("第1层任务开始")

		err := pool.WaitGO(func() error {
			atomic.AddInt32(&taskCount, 1)
			t.Log("  第2层任务开始")

			err := pool.WaitGOIndex(5, func(index int) error {
				atomic.AddInt32(&taskCount, 1)
				t.Logf("    第3层任务 %d 开始", index)

				err := pool.WaitGO(func() error {
					atomic.AddInt32(&taskCount, 1)
					t.Log("      第4层任务执行")
					time.Sleep(2 * time.Second)
					return nil
				})
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Errorf("执行出错: %v", err)
		return
	}
	
	count := atomic.LoadInt32(&taskCount)
	expected := int32(12)
	t.Logf("总任务数: %d, 预期: %d", count, expected)
	if count != expected {
		t.Errorf("任务数量不匹配: 期望 %d, 实际 %d", expected, count)
	}
}
