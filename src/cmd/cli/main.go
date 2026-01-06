package main

import (
	"context"
	"fmt"
	"log"
	"myobj/src/config"
	"myobj/src/internal/repository/database"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"myobj/src/pkg/util"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

const (
	AppName    = "MyObj CLI"
	AppVersion = "1.0.0"
)

var (
	db         *impl.RepositoryFactory
	cacheStore cache.Cache
)

func main() {
	app := &cli.App{
		Name:    AppName,
		Version: AppVersion,
		Usage:   "MyObj 系统管理工具",
		Before:  initialize,
		Commands: []*cli.Command{
			{
				Name:    "user",
				Aliases: []string{"u"},
				Usage:   "用户管理",
				Subcommands: []*cli.Command{
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Usage:   "列出所有用户",
						Action:  listUsersAction,
					},
					{
						Name:      "detail",
						Aliases:   []string{"info"},
						Usage:     "查看用户详情",
						ArgsUsage: "<username>",
						Action:    userDetailAction,
					},
					{
						Name:      "reset-password",
						Aliases:   []string{"resetpwd", "pwd"},
						Usage:     "重置用户密码",
						ArgsUsage: "<username> <new-password>",
						Action:    resetPasswordAction,
					},
					{
						Name:      "change-group",
						Aliases:   []string{"chgrp"},
						Usage:     "修改用户组（交互式选择）",
						ArgsUsage: "<username>",
						Action:    changeGroupAction,
					},
					{
						Name:      "ban",
						Usage:     "封禁用户",
						ArgsUsage: "<username>",
						Action:    banUserAction,
					},
					{
						Name:      "unban",
						Usage:     "解封用户",
						ArgsUsage: "<username>",
						Action:    unbanUserAction,
					},
					{
						Name:      "kick",
						Usage:     "踢出用户登录会话",
						ArgsUsage: "<username>",
						Action:    kickUserAction,
					},
				},
			},
			{
				Name:    "group",
				Aliases: []string{"g"},
				Usage:   "组管理",
				Subcommands: []*cli.Command{
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Usage:   "列出所有用户组",
						Action:  listGroupsAction,
					},
				},
			},
			{
				Name:    "system",
				Aliases: []string{"sys"},
				Usage:   "系统信息",
				Subcommands: []*cli.Command{
					{
						Name:   "info",
						Usage:  "查看系统信息",
						Action: systemInfoAction,
					},
					{
						Name:   "stats",
						Usage:  "查看系统统计",
						Action: systemStatsAction,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		pterm.Error.Println(err)
		os.Exit(1)
	}
}

// initialize 初始化系统组件
func initialize(c *cli.Context) error {
	pterm.DefaultHeader.WithFullWidth().Println("MyObj CLI 管理工具")
	pterm.Info.Println("正在初始化...")

	// 加载配置
	if err := loadConfig(); err != nil {
		return fmt.Errorf("配置加载失败: %w", err)
	}

	// 初始化日志
	if err := setupLogger(); err != nil {
		return fmt.Errorf("日志初始化失败: %w", err)
	}

	// 初始化数据库
	if err := setupDatabase(); err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	// 初始化缓存
	cacheStore = cache.InitCache()

	pterm.Success.Println("初始化完成")
	fmt.Println()
	return nil
}

func loadConfig() error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[错误] 配置加载异常: %v\n", r)
		}
	}()
	config.InitConfig()
	return nil
}

func setupLogger() error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[错误] 日志系统初始化异常: %v\n", r)
		}
	}()
	logger.InitLogger()
	return nil
}

func setupDatabase() error {
	defer func() {
		if r := recover(); r != nil {
			logger.LOG.Error("[错误] 数据库连接异常", "panic", r)
		}
	}()
	database.InitDataBase()
	db = impl.NewRepositoryFactory(database.GetDB())
	return nil
}

// ========== 用户管理命令 ==========

// listUsersAction 列出所有用户
func listUsersAction(c *cli.Context) error {
	ctx := context.Background()

	spinner, _ := pterm.DefaultSpinner.Start("正在获取用户列表...")
	users, err := db.User().List(ctx, 0, 10000)
	spinner.Stop()

	if err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}

	if len(users) == 0 {
		pterm.Warning.Println("暂无用户")
		return nil
	}

	// 构建表格数据
	tableData := pterm.TableData{
		{"ID", "用户名", "昵称", "邮箱", "组ID", "状态", "空间(GB)", "创建时间"},
	}

	for _, user := range users {
		status := "正常"
		if user.State == 1 {
			status = pterm.Red("封禁")
		} else {
			status = pterm.Green("正常")
		}

		spaceGB := fmt.Sprintf("%.2f", float64(user.Space)/1024/1024/1024)
		tableData = append(tableData, []string{
			user.ID[:8] + "...",
			user.UserName,
			user.Name,
			user.Email,
			fmt.Sprintf("%d", user.GroupID),
			status,
			spaceGB,
			user.CreatedAt.Format("2006-01-02"),
		})
	}

	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	pterm.Info.Printf("共 %d 个用户\n", len(users))
	return nil
}

// userDetailAction 查看用户详情
func userDetailAction(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("请指定用户名")
	}

	username := c.Args().Get(0)
	ctx := context.Background()

	user, err := db.User().GetByUserName(ctx, username)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	// 获取组信息
	group, _ := db.Group().GetByID(ctx, user.GroupID)
	groupName := "未知"
	if group != nil {
		groupName = group.Name
	}

	status := "正常"
	statusColor := pterm.Green
	if user.State == 1 {
		status = "封禁"
		statusColor = pterm.Red
	}

	// 显示用户详情
	pterm.DefaultSection.Println("用户详情")
	fmt.Println()

	details := [][]string{
		{"用户ID", user.ID},
		{"用户名", user.UserName},
		{"昵称", user.Name},
		{"邮箱", user.Email},
		{"手机", user.Phone},
		{"用户组", fmt.Sprintf("%s (ID: %d)", groupName, user.GroupID)},
		{"状态", statusColor(status)},
		{"总空间", fmt.Sprintf("%.2f GB", float64(user.Space)/1024/1024/1024)},
		{"剩余空间", fmt.Sprintf("%.2f GB", float64(user.FreeSpace)/1024/1024/1024)},
		{"创建时间", user.CreatedAt.Format("2006-01-02 15:04:05")},
	}

	for _, detail := range details {
		pterm.Printf("  %s: %s\n", pterm.Bold.Sprint(detail[0]), detail[1])
	}

	return nil
}

// resetPasswordAction 重置用户密码
func resetPasswordAction(c *cli.Context) error {
	if c.NArg() < 2 {
		return fmt.Errorf("用法: reset-password <username> <new-password>")
	}

	username := c.Args().Get(0)
	newPassword := c.Args().Get(1)

	if len(newPassword) < 6 {
		return fmt.Errorf("密码长度不能少于6位")
	}

	ctx := context.Background()

	// 查询用户
	user, err := db.User().GetByUserName(ctx, username)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	// 确认操作
	confirm := false
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("确定要重置用户 '%s' 的密码吗？", username),
	}
	if err := survey.AskOne(prompt, &confirm); err != nil {
		return err
	}

	if !confirm {
		pterm.Info.Println("操作已取消")
		return nil
	}

	// Hash密码
	hashedPassword, err := util.GeneratePassword(newPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	user.Password = hashedPassword

	// 更新数据库
	if err := db.User().Update(ctx, user); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	pterm.Success.Printf("用户 '%s' 的密码已重置\n", username)
	return nil
}

// changeGroupAction 修改用户组（交互式）
func changeGroupAction(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("请指定用户名")
	}

	username := c.Args().Get(0)
	ctx := context.Background()

	// 查询用户
	user, err := db.User().GetByUserName(ctx, username)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	// 获取所有组
	groups, err := db.Group().List(ctx, 0, 1000)
	if err != nil {
		return fmt.Errorf("获取组列表失败: %w", err)
	}

	if len(groups) == 0 {
		return fmt.Errorf("系统中没有可用的用户组")
	}

	// 构建选项
	groupOptions := make([]string, len(groups))
	groupMap := make(map[string]*models.Group)

	for i, group := range groups {
		option := fmt.Sprintf("%s (ID:%d)", group.Name, group.ID)
		groupOptions[i] = option
		groupMap[option] = group
	}

	// 交互式选择
	var selectedOption string
	prompt := &survey.Select{
		Message: fmt.Sprintf("请选择用户 '%s' 的新用户组:", username),
		Options: groupOptions,
	}

	if err := survey.AskOne(prompt, &selectedOption); err != nil {
		return err
	}

	selectedGroup := groupMap[selectedOption]

	// 确认操作
	confirm := false
	confirmPrompt := &survey.Confirm{
		Message: fmt.Sprintf("确定将用户 '%s' 从组 %d 改为 '%s' (ID:%d) 吗？",
			username, user.GroupID, selectedGroup.Name, selectedGroup.ID),
	}
	if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
		return err
	}

	if !confirm {
		pterm.Info.Println("操作已取消")
		return nil
	}

	// 更新用户组
	oldGroupID := user.GroupID
	user.GroupID = selectedGroup.ID

	if err := db.User().Update(ctx, user); err != nil {
		return fmt.Errorf("更新用户组失败: %w", err)
	}

	pterm.Success.Printf("用户 '%s' 已从组 %d 变更为 '%s' (ID:%d)\n",
		username, oldGroupID, selectedGroup.Name, selectedGroup.ID)
	return nil
}

// banUserAction 封禁用户
func banUserAction(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("请指定用户名")
	}

	username := c.Args().Get(0)
	ctx := context.Background()

	user, err := db.User().GetByUserName(ctx, username)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	if user.State == 1 {
		pterm.Warning.Printf("用户 '%s' 已经处于封禁状态\n", username)
		return nil
	}

	// 确认操作
	confirm := false
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("确定要封禁用户 '%s' 吗？", username),
	}
	if err := survey.AskOne(prompt, &confirm); err != nil {
		return err
	}

	if !confirm {
		pterm.Info.Println("操作已取消")
		return nil
	}

	user.State = 1

	if err := db.User().Update(ctx, user); err != nil {
		return fmt.Errorf("封禁用户失败: %w", err)
	}

	pterm.Success.Printf("用户 '%s' 已被封禁\n", username)
	return nil
}

// unbanUserAction 解封用户
func unbanUserAction(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("请指定用户名")
	}

	username := c.Args().Get(0)
	ctx := context.Background()

	user, err := db.User().GetByUserName(ctx, username)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	if user.State == 0 {
		pterm.Warning.Printf("用户 '%s' 当前未被封禁\n", username)
		return nil
	}

	// 确认操作
	confirm := false
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("确定要解封用户 '%s' 吗？", username),
	}
	if err := survey.AskOne(prompt, &confirm); err != nil {
		return err
	}

	if !confirm {
		pterm.Info.Println("操作已取消")
		return nil
	}

	user.State = 0

	if err := db.User().Update(ctx, user); err != nil {
		return fmt.Errorf("解封用户失败: %w", err)
	}

	pterm.Success.Printf("用户 '%s' 已解封\n", username)
	return nil
}

// kickUserAction 踢出用户登录
func kickUserAction(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("请指定用户名")
	}

	username := c.Args().Get(0)
	ctx := context.Background()

	user, err := db.User().GetByUserName(ctx, username)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	// 确认操作
	confirm := false
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("确定要踢出用户 '%s' 的所有登录会话吗？", username),
	}
	if err := survey.AskOne(prompt, &confirm); err != nil {
		return err
	}

	if !confirm {
		pterm.Info.Println("操作已取消")
		return nil
	}

	// 清除缓存中与该用户相关的JWT tokens
	// 由于缓存中key是token本身，我们需要清除所有缓存（或使用特定前缀）
	// 这里采用简单策略：清除所有缓存
	cacheStore.Clear()

	pterm.Success.Printf("用户 '%s' (ID: %s) 的所有登录会话已被清除\n", username, user.ID)
	pterm.Info.Println("注意：为确保完全清除，已清空所有缓存")
	return nil
}

// ========== 组管理命令 ==========

// listGroupsAction 列出所有组
func listGroupsAction(c *cli.Context) error {
	ctx := context.Background()

	spinner, _ := pterm.DefaultSpinner.Start("正在获取组列表...")
	groups, err := db.Group().List(ctx, 0, 1000)
	spinner.Stop()

	if err != nil {
		return fmt.Errorf("查询组失败: %w", err)
	}

	if len(groups) == 0 {
		pterm.Warning.Println("暂无用户组")
		return nil
	}

	// 统计每个组的用户数
	users, _ := db.User().List(ctx, 0, 10000)
	groupUserCount := make(map[int]int)
	for _, user := range users {
		groupUserCount[user.GroupID]++
	}

	// 构建表格
	tableData := pterm.TableData{
		{"ID", "组名", "默认组", "空间(GB)", "用户数", "创建时间"},
	}

	for _, group := range groups {
		isDefault := "否"
		if group.GroupDefault == 1 {
			isDefault = pterm.Green("是")
		}

		spaceGB := fmt.Sprintf("%.2f", float64(group.Space)/1024/1024/1024)
		userCount := groupUserCount[group.ID]

		tableData = append(tableData, []string{
			fmt.Sprintf("%d", group.ID),
			group.Name,
			isDefault,
			spaceGB,
			fmt.Sprintf("%d", userCount),
			group.CreatedAt.Format("2006-01-02"),
		})
	}

	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	pterm.Info.Printf("共 %d 个用户组\n", len(groups))
	return nil
}

// ========== 系统信息命令 ==========

// systemInfoAction 查看系统信息
func systemInfoAction(c *cli.Context) error {
	pterm.DefaultSection.Println("系统信息")
	fmt.Println()

	cfg := config.GetConfig()

	info := [][]string{
		{"数据库类型", strings.ToUpper(cfg.Database.Type)},
		{"缓存类型", strings.ToUpper(cfg.Cache.Type)},
		{"应用名称", AppName},
		{"应用版本", AppVersion},
	}

	for _, item := range info {
		pterm.Printf("  %s: %s\n", pterm.Bold.Sprint(item[0]), item[1])
	}

	return nil
}

// systemStatsAction 查看系统统计
func systemStatsAction(c *cli.Context) error {
	ctx := context.Background()

	spinner, _ := pterm.DefaultSpinner.Start("正在统计系统数据...")

	// 统计用户数
	totalUsers, _ := db.User().Count(ctx)
	users, _ := db.User().List(ctx, 0, 10000)
	activeUsers := 0
	bannedUsers := 0
	for _, user := range users {
		if user.State == 0 {
			activeUsers++
		} else {
			bannedUsers++
		}
	}

	// 统计组数
	totalGroups, _ := db.Group().Count(ctx)

	spinner.Stop()

	pterm.DefaultSection.Println("系统统计")
	fmt.Println()

	stats := [][]string{
		{"总用户数", fmt.Sprintf("%d", totalUsers)},
		{"正常用户", pterm.Green(fmt.Sprintf("%d", activeUsers))},
		{"封禁用户", pterm.Red(fmt.Sprintf("%d", bannedUsers))},
		{"用户组数", fmt.Sprintf("%d", totalGroups)},
	}

	for _, stat := range stats {
		pterm.Printf("  %s: %s\n", pterm.Bold.Sprint(stat[0]), stat[1])
	}

	return nil
}
