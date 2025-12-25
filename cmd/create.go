/*
Package cmd
Copyright © 2025 getcharzp <getcharzp@gmail.com>
*/
package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/up-zero/gotool/convertutil"
	proxyClient "github.com/up-zero/my-proxy/client/proxy"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/service/proxy"
)

var (
	proxyName  string
	proxyType  string
	listenAddr string
	listenPort string
	targetAddr string
	targetPort string
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "create a new proxy service",
	Run: func(cmd *cobra.Command, args []string) {
		pb := &models.ProxyBasic{
			Name:          proxyName,
			Type:          proxyType,
			ListenAddress: listenAddr,
			ListenPort:    listenPort,
			TargetAddress: targetAddr,
			TargetPort:    targetPort,
		}
		if len(args) > 0 {
			pb.Name = args[0]
		}

		// 代理配置完整，直接创建
		if pb.Name != "" && pb.Type != "" && pb.ListenPort != "" &&
			pb.TargetAddress != "" && pb.TargetPort != "" {
			req := new(proxy.CreateRequest)
			convertutil.CopyProperties(pb, req)
			if err := proxyClient.Create(req); err != nil {
				fmt.Printf("proxy create faile: %v\n", err)
				return
			}
			fmt.Println("proxy create success")
			return
		}

		// 代理配置不完整，使用 TUI 创建
		p := tea.NewProgram(initialModel(pb))
		if _, err := p.Run(); err != nil {
			fmt.Printf("TUI run error: %v\n", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVar(&proxyName, "name", "", "proxy name")
	createCmd.Flags().StringVar(&proxyType, "type", "", "proxy type")
	createCmd.Flags().StringVar(&listenAddr, "laddr", "", "listen address")
	createCmd.Flags().StringVar(&listenPort, "lport", "", "listen port")
	createCmd.Flags().StringVar(&targetAddr, "taddr", "", "target address")
	createCmd.Flags().StringVar(&targetPort, "tport", "", "target port")
}
