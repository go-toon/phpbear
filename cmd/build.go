package cmd

import (
	"github.com/urfave/cli"
	"fmt"
	"github.com/go-toon/phpbear/git"
	"path/filepath"
	"io/ioutil"
	"encoding/json"
	"os"
	"strings"
)

var (
	CmdBuild = cli.Command{
		Name: "build",
		Usage: "构建php项目",
		Description: "自动化构建项目",
		Action: runBuild,
		Flags: []cli.Flag{
			stringFlag("project-git, p", "", "项目git url地址"),
			stringFlag("project-alias, a", "", "项目Alias名称"),
			stringFlag("project-branch, b", "master", "项目源分支名称"),
			stringFlag("feature-branch, f", "", "项目功能分支名称， 若不存在，则自动创建"),
		},

	}
)

func runBuild(c *cli.Context) error {
	if !c.IsSet("project-git") {
		fmt.Print(" 请输入项目Git地址: ")
		if _, err := fmt.Scanf("%s", &ProjectGit); err != nil {
			return cli.NewExitError("项目Git地址获取失败", 86)
		}
	} else {
		ProjectGit = c.String("project-git")
	}

	ProjectTag = c.String("project-branch")
	ProjectAlias = c.String("project-alias")
	FeatureBranch = c.String("feature-branch")
	ProjectName = getProjectName(ProjectGit)

	if ProjectAlias == "" {
		ProjectAlias = ProjectName
	}

	// 克隆项目代码
	fmt.Printf("开始clone项目：%s\n", ProjectGit)
	if _, err := git.NewCommand("clone", "-b", ProjectTag, ProjectGit, ProjectAlias).Run(); err != nil {
		return cli.NewExitError(err.Error(), 86)
	}

	// 获取项目目录的绝对路径
	projectPath, _ := filepath.Abs("./" + ProjectAlias)

	// 创建新分支
	if FeatureBranch != "" && FeatureBranch != ProjectTag {
		_, err := git.NewCommand("checkout", "-b", FeatureBranch).RunInDir(projectPath)
		if err != nil {
			return fmt.Errorf("项目切换分支失败：%v", err)
		}
	}

	if err := buildProject(ProjectAlias); err != nil {
		//构建失败，删除目录
		os.Remove(ProjectAlias)
		return cli.NewExitError(err.Error(), 86)
	}

	// 执行composer install工作
	fmt.Println("composer install start ...")
	if _, err := git.NewComposerCommand("install").RunInDir(projectPath); err != nil {
		return fmt.Errorf("composer install 失败，请自行执行命令: %v", err)
	}
	fmt.Println("composer install end ...")

	return nil
}

func buildProject(projectPath string) error {
	relativePath := "./" + projectPath

	// 检测phpbear.json文件是否存在
	if !isFile(relativePath + "/phpbear.json") {
		return fmt.Errorf("项目中缺少构建文件：phpbear.json")
	}

	jsonBlob, err := ioutil.ReadFile(relativePath + "/phpbear.json")
	if err != nil {
		return fmt.Errorf("读取phpbear.json配置文件失败: %v", err)
	}

	var bearOption PhpBearOptions
	err = json.Unmarshal(jsonBlob, &bearOption)
	if err != nil {
		return fmt.Errorf("读取phpbear.json 配置文件失败: %v", err)
	}

	// clone标准框架至tmp目录
	fmt.Println("开始clone framework...")
	if _, err := git.NewCommand("clone", "-b", bearOption.FrameworkTag, bearOption.FrameworkGit, projectPath+"@tmp").Run(); err != nil {
		return cli.NewExitError(err.Error(), 86)
	}

	fmt.Println("合并代码 start...")

	if err := mergeFrameToProject(projectPath+"@tmp", projectPath); err != nil {
		return cli.NewExitError(err.Error(), 86)
	}

	// 删除tmp目录
	if err := os.RemoveAll(projectPath+"@tmp"); err != nil {
		return fmt.Errorf("删除临时目录失败： %v", err)
	}

	fmt.Println("合并代码 end ...")

	return nil
}

func mergeFrameToProject(framePath, projectPath string) error {
	//首先删除framePath中的.git目录
	os.RemoveAll("./" + framePath + "/.git")

	var mergeFunc filepath.WalkFunc = func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return err
		}
		//var tmpPath = strings.Trim(path, framePath)
		var tmpPath = strings.TrimPrefix(path, framePath)

		if path == framePath {
			// 根目录跳过
			return nil
		}

		if info.IsDir() {
			// 检测该目录在projectPath中是否存在，不存在则创建
			if !isDir("./" + projectPath + "/" + tmpPath) {
				os.MkdirAll("./" + projectPath + "/" + tmpPath, 0766)
			}

			return nil
		}

		// 文件，检查文件在项目目录中是否存在，不存在则复制文件
		if !isExist("./" + projectPath + "/" + tmpPath) {
			copyFile(path, "./" + projectPath + "/" + tmpPath)
		}

		return err
	}

	if err := filepath.Walk(framePath, mergeFunc); err != nil {
		return err
	}

	// 合并产生.gitignore文件
	gitignoreData := make([]byte, 1024)
	if isExist("./" + projectPath + "/.gitignore_project") {
		fi, _ := os.Open("./" + projectPath + "/.gitignore_project")
		defer fi.Close()

		gitignoreData, _ = ioutil.ReadAll(fi)
	}

	if isExist("./" + projectPath + "/.gitignore_framework") {
		fi, _ := os.Open("./" + projectPath + "/.gitignore_framework")
		defer fi.Close()

		gitignoreData, _ = ioutil.ReadAll(fi)
	}

	fi, _ := os.OpenFile("./" + projectPath + "/.gitignore", os.O_RDWR|os.O_CREATE, 0755)
	defer fi.Close()
	fi.Write(gitignoreData)

	return nil
}