package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"

	"gorgonia.org/gorgonia"
)

// 使用莱布尼茨级数，计算圆周率Pi
// Pi/4=1-1/3+1/5-1/7+1/9-...
// ∞
// ∑ (−1)^n * 1/(2n+1)
// n=0
func main() {
	g := gorgonia.NewGraph()
	var pi, four, threshold *gorgonia.Node
	var thresholdV float64
	var err error

	// 锁定 OS 线程以启用 CUDA
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// 定义π的初始值
	pi = gorgonia.NewScalar(g, gorgonia.Float64, gorgonia.WithName("pi"))
	gorgonia.Let(pi, float64(0)) // 初始化π为0

	// 定义四倍的值
	four = gorgonia.NewScalar(g, gorgonia.Float64, gorgonia.WithName("four"))
	gorgonia.Let(four, 4.0)

	// 定义精度阈值
	thresholdEnv := os.Getenv("THRESHOLD")
	if thresholdEnv != "" {
		// 将字符串转换为float64
		thresholdV, err = strconv.ParseFloat(thresholdEnv, 64)
		if err != nil {
			fmt.Println("Error converting THRESHOLD ENV to float64:", err)
			return
		}
	}

	threshold = gorgonia.NewScalar(g, gorgonia.Float64, gorgonia.WithName("threshold"))
	// 假设1e-6为所需的精度
	if thresholdV != float64(0) {
		gorgonia.Let(threshold, thresholdV)
	} else {
		gorgonia.Let(threshold, 1e-1)
	}

	// 循环计算莱布尼茨级数的每一项
	for n := 0; ; n++ {
		var numerator, denominator, sign, item *gorgonia.Node
		sign = gorgonia.NewScalar(g, gorgonia.Float64, gorgonia.WithName(fmt.Sprintf("sign%v", n)))
		// 交替正负号
		if n%2 != 0 {
			gorgonia.Let(sign, float64(-1))
		} else {
			gorgonia.Let(sign, float64(1))
		}

		denominator = gorgonia.NewScalar(g, gorgonia.Float64, gorgonia.WithName(fmt.Sprintf("denominator%v", n)))
		// 分母为奇数
		gorgonia.Let(denominator, float64(2*n+1))

		numerator = gorgonia.NewScalar(g, gorgonia.Float64, gorgonia.WithName(fmt.Sprintf("numerator%v", n)))
		// 分子为1
		gorgonia.Let(numerator, float64(1.0))

		// 计算当前项1/(2n+1)
		item, err = gorgonia.Div(numerator, denominator)
		if err != nil {
			log.Fatal(err)
		}

		// 应用正负号
		item, err = gorgonia.Mul(sign, item)
		if err != nil {
			log.Fatal(err)
		}

		// 累加当前项到π
		pi, err = gorgonia.Add(pi, item)
		if err != nil {
			log.Fatal(err)
		}

		// 检查是否达到所需的精度
		item, err := gorgonia.Abs(item)
		if err != nil {
			log.Fatal(err)
		}

		// 创建虚拟机，用于执行图，并使用GPU参与
		machine := gorgonia.NewTapeMachine(g, gorgonia.UseCudaFor("add", "mul", "div"))
		defer machine.Close()

		if err = machine.RunAll(); err != nil {
			log.Fatal(err)
		}

		// fmt.Printf("item %v", item.Value().Data().(float64))
		if item.Value().Data().(float64) < threshold.Value().Data().(float64) {
			break
		}
	}

	// 计算最终的π值, π=4 * (π/4)
	pi, err = gorgonia.Mul(pi, four)
	if err != nil {
		log.Fatal(err)
	}
	// 创建虚拟机，用于执行图，并使用GPU参与
	machine := gorgonia.NewTapeMachine(g, gorgonia.UseCudaFor("add", "mul", "div"))
	defer machine.Close()
	if err = machine.RunAll(); err != nil {
		log.Fatal(err)
	}

	// 输出π的近似值
	fmt.Printf("Approximated value of π: %v\n", pi.Value())
}
