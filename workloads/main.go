package main

import (
	"fmt"
	"log"
	"runtime"

	"gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
	"k8s.io/klog"
)

func main() {
	g := gorgonia.NewGraph()
	var x, y, z *gorgonia.Node
	var err error

	// 创建两个多维矩阵节点
	x = gorgonia.NewMatrix(g, tensor.Float64, gorgonia.WithShape(2, 2), gorgonia.WithName("x"))
	y = gorgonia.NewMatrix(g, tensor.Float64, gorgonia.WithShape(2, 2), gorgonia.WithName("y"))

	// 定义矩阵乘法的表达式
	z, err = gorgonia.Mul(x, y)
	if err != nil {
		log.Fatal(err)
	}

	// 创建一个虚拟机来运行计算图，指定使用 GPU
	m := gorgonia.NewTapeMachine(g, gorgonia.UseCudaFor("mul"))
	defer m.Close()

	// 初始化矩阵值
	xData := tensor.New(tensor.WithShape(2, 2), tensor.WithBacking([]float64{1, 2, 3, 4}))
	yData := tensor.New(tensor.WithShape(2, 2), tensor.WithBacking([]float64{5, 6, 7, 8}))

	// 将数据赋值给节点
	gorgonia.Let(x, xData)
	gorgonia.Let(y, yData)

	// 锁定 OS 线程以启用 CUDA
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// 运行计算图
	if err = m.RunAll(); err != nil {
		klog.Info(err)
		log.Fatal(err)
	}

	// 获取计算结果
	result := z.Value()

	fmt.Printf("Result of matrix multiplication:\n%v\n", result)
}
