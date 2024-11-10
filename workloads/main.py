import torch

def estimate_pi(threshold):
    # 检查CUDA是否可用
    if torch.cuda.is_available():
        print("CUDA is available! Training on GPU.")
        device = torch.device("cuda")
    else:
        print("CUDA is not available. Training on CPU.")
        device = torch.device("cpu")

    # 在CUDA设备上生成随机点
    x = torch.rand(threshold, device=device)
    y = torch.rand(threshold, device=device)
    
    # 计算点到原点的距离
    distance = torch.sqrt(x**2 + y**2)
    
    # 计算落在圆内的点的数量
    inside_circle = (distance <= 1).sum().item()
    
    # 估计圆周率
    pi_estimate = 4 * inside_circle / threshold
    return pi_estimate

# 使用1亿个样本来估计圆周率
pi_estimate = estimate_pi(100000000)
print(f"Estimated π: {pi_estimate}")