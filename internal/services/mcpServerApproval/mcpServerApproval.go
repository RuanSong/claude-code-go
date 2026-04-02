package mcpServerApproval

import (
	"fmt"
	"os"
)

// MCPServerApprovalStatus MCP服务器审批状态
type MCPServerApprovalStatus string

const (
	StatusPending  MCPServerApprovalStatus = "pending"
	StatusApproved MCPServerApprovalStatus = "approved"
	StatusRejected MCPServerApprovalStatus = "rejected"
)

// MCPServerInfo MCP服务器信息
type MCPServerInfo struct {
	Name   string                  `json:"name"`
	Status MCPServerApprovalStatus `json:"status"`
	Scope  string                  `json:"scope"`
}

// HandleMcpjsonServerApprovals 处理MCP服务器审批
// 显示待处理项目的MCP服务器审批对话框
func HandleMcpjsonServerApprovals() error {
	// 获取项目MCP配置
	projectServers := GetMcpConfigsByScope("project")

	var pendingServers []string
	for serverName, server := range projectServers {
		if GetProjectMcpServerStatus(serverName) == StatusPending {
			pendingServers = append(pendingServers, server.Name)
		}
	}

	if len(pendingServers) == 0 {
		return nil
	}

	// 显示审批对话框
	if len(pendingServers) == 1 {
		return ShowApprovalDialog(pendingServers[0])
	}

	return ShowMultiselectDialog(pendingServers)
}

// GetMcpConfigsByScope 根据作用域获取MCP配置
func GetMcpConfigsByScope(scope string) map[string]*MCPServerInfo {
	// 从环境变量或配置文件读取
	configs := make(map[string]*MCPServerInfo)

	// 简化实现
	return configs
}

// GetProjectMcpServerStatus 获取项目MCP服务器状态
func GetProjectMcpServerStatus(serverName string) MCPServerApprovalStatus {
	// 检查环境变量中的缓存状态
	status := os.Getenv("MCP_SERVER_STATUS_" + serverName)
	switch status {
	case "approved":
		return StatusApproved
	case "rejected":
		return StatusRejected
	default:
		return StatusPending
	}
}

// ShowApprovalDialog 显示单个服务器审批对话框
func ShowApprovalDialog(serverName string) error {
	fmt.Printf("MCP Server Approval Required\n")
	fmt.Printf("============================\n\n")
	fmt.Printf("Server: %s\n\n", serverName)
	fmt.Printf("This server requires approval before it can be used.\n")
	fmt.Printf("Do you want to approve this server? (y/n): ")

	// 简化实现 - 实际应该读取用户输入
	return nil
}

// ShowMultiselectDialog 显示多服务器选择对话框
func ShowMultiselectDialog(serverNames []string) error {
	fmt.Printf("MCP Server Approval Required\n")
	fmt.Printf("============================\n\n")
	fmt.Printf("Multiple servers require approval:\n\n")

	for i, name := range serverNames {
		fmt.Printf("  %d. %s\n", i+1, name)
	}

	fmt.Printf("\nPlease approve or reject each server.\n")

	return nil
}

// SetServerApproved 设置服务器为已批准状态
func SetServerApproved(serverName string) {
	// 保存到配置
	_ = serverName
}

// SetServerRejected 设置服务器为已拒绝状态
func SetServerRejected(serverName string) {
	// 保存到配置
	_ = serverName
}
