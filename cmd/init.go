package cmd

import (
	"github.com/urfave/cli"
	"fmt"
	"path/filepath"
	"os"
	"github.com/go-toon/phpbear/git"
	"io/ioutil"
	"time"
	"encoding/json"
)

var (
	CmdInit = cli.Command{
		Name: "init",
		Usage: "初始化项目",
		Description: "自动初始化项目",
		Action: runInit,
		Flags: []cli.Flag{
			stringFlag("framework-git, f", "http://root@github.com/PHP/vender/Yaf.git", "The git remote url of Yaf-Framework "),
			stringFlag("framework-tag, t", "master", "The branch or tag of the Yaf-framework" ),
			stringFlag("project-git, p", "", "The git url of the project which will be created" ),
			stringFlag("project-alias, a", "", "The dir of the project"),
		},
	}
)

func runInit(c *cli.Context) error {
	if !c.IsSet("project-git") {
		fmt.Print("请输入项目Git地址: ")

		if _, err := fmt.Scanf("%s", &ProjectGit); err != nil {
			return cli.NewExitError("项目Git地址获取失败", 86)
		}
		// return fmt.Errorf("Project-git is not specified")
	} else {
		ProjectGit = c.String("project-git")
	}

	FrameworkGit = c.String("framework-git")
	FrameworkTag = c.String("framework-tag")
	ProjectAlias = c.String("project-alias")
	ProjectName  = getProjectName(ProjectGit)

	if ProjectAlias == "" {
		ProjectAlias = ProjectName
	}

	// 克隆项目代码
	fmt.Printf("开始clone项目：%s\n", ProjectGit)
	if _, err := git.NewCommand("clone", ProjectGit, ProjectAlias).Run(); err != nil {
		return cli.NewExitError(err.Error(), 86)
	}

	// 开始构建项目框架
	if err := initProject(ProjectAlias); err != nil {
		os.Remove(ProjectAlias)
		return cli.NewExitError(err.Error(), 86)
	}


	// git add commit
	projectPath, _ := filepath.Abs("./" + ProjectAlias)
	if err := git.AddChanges(projectPath, true); err != nil {
		return fmt.Errorf("git Add Changes: %v", err)
	} else if err := git.CommitChanges(projectPath, "phpbear 初始化项目"); err != nil  {
		return fmt.Errorf("git Push: %v", err)
	} else if err = git.Push(projectPath, "origin", "master"); err != nil {
		return fmt.Errorf("git Push: %v", err)
	}

	return nil
}

func initProject(projectPath string) error {
	relativePath := "./" + projectPath
	// 判断application目录是否存在
	if ! isDir(relativePath + "/application") {
		if err := os.MkdirAll(relativePath + "/application/controllers", 0766); err != nil {
			return fmt.Errorf("create dir /application/controllers: %v", err)
		}
		if err := os.MkdirAll(relativePath + "/application/library", 0766); err != nil {
			return fmt.Errorf("create dir /application/library: %v", err)
		}
		if err := os.MkdirAll(relativePath + "/application/models", 0766); err != nil {
			return fmt.Errorf("create dir /application/models: %v", err)
		}

		// 创建application/controllers/Index.php, Error.php
		indexStr := []byte(`<?php
/**
* Index 控制器
* @author phpbear
*/
class IndexController extends Yaf\Controller_Abstract
{
    public function indexAction()
    {
        echo "hello world";
    }
}
		`)
		err := ioutil.WriteFile(relativePath + "/application/controllers/Index.php", indexStr, 0644)
		if err != nil {
			return fmt.Errorf("create file /application/controllers/Index.php: %v", err)
		}

		errorStr := []byte(`<?php
/**
 * @name ErrorController
 * @desc 错误控制器, 在发生未捕获的异常时刻被调用
 * @author phpbear
 */
use Yaf\Controller_Abstract;
class ErrorController extends Controller_Abstract {

	//从2.1开始, errorAction支持直接通过参数获取异常
	public function errorAction(Exception $exception) {
        $errMsg = $this->_makePrettyException($exception) . $exception->getTraceAsString();
        try {
            throw $exception;
        } catch (Yaf\Exception_LoadFailed $e) {

            header('HTTP/1.1 404 Not Found');
            header("status: 404 Not Found");

            exit;
        } catch (Yaf\Exception $e) {

            header('HTTP/1.1 500 Internal Server Error');
            exit;
        }
	}

    private function _makePrettyException(Exception $e)
    {
        $trace = $e->getTrace();
        $result = 'Exception: "';
        $result .= $e->getMessage();
        $result .= '" @ ';
        if($trace[0]['class'] != '') {
            $result .= $trace[0]['class'];
            $result .= '->';
        }
        $result .= $trace[0]['function'];
        $result .= "();\n";

        return $result;
    }
}

		`)

		err = ioutil.WriteFile(relativePath + "/application/controllers/Error.php", errorStr, 0644)
		if err != nil {
			return fmt.Errorf("create file /application/controllers/Error.php: %v", err)
		}

		// 在library及models下创建空文件.gitkeep
		if _, err = os.Create(relativePath + "/application/library/.gitkeep"); err != nil {
			return fmt.Errorf("create file .gitkeep error: %v", err)
		}

		if _, err = os.Create(relativePath + "/application/models/.gitkeep"); err != nil {
			return fmt.Errorf("create file .gitkeep error : %v", err)
		}

	}

	if !isDir(relativePath + "/conf") {
		if err := os.MkdirAll(relativePath + "/conf", 0766); err != nil {
			return err
		}

		// 创建application.ini文件
		appIniStr := []byte(`[common]
application.directory = APPLICATION_PATH  "/application"
application.dispatcher.catchException = TRUE

;;;disconf配置
disconf.url = http://disconf.51gaga.cn:8081/api/config/file
disconf.key = app.ini
disconf.version = 1.0.0
disconf.app = composerDemo
disconf.type = 0
disconf.env = rd
		`)

		err := ioutil.WriteFile(relativePath + "/conf/application.ini", appIniStr, 0644)
		if err != nil {
			return err
		}
	}

	//创建.gitignore_project文件
	if !isFile(relativePath + "/.gitignore_project") {
		gitIngoreStr := []byte(`.idea
.gitignore
.gitignore_framework
.DS_Store
/composer.lock
vendor
.buildpath
.project
	    `)
		err := ioutil.WriteFile(relativePath + "/.gitignore_project", gitIngoreStr, 0644)
		if err != nil {
			return err
		}
	}

	//创建composer.json文件
	if !isFile(relativePath + "/composer.json") {
		composerStr := []byte(`{
    "name": "51gaga/php-demo",
    "description": "php-demo",
    "type": "product",
    "license": "MIT",
    "require": {
        "51gaga/utils":"*",
        "51gaga/log":"*",
        "51gaga/db": "*",
        "51gaga/curl":"*",
        "51gaga/toon":"*",
        "51gaga/html":"*"
    }
}
		`)
		err := ioutil.WriteFile(relativePath + "/composer.json", composerStr, 0644)
		if err != nil {
			return err
		}
	}

	//创建phpbear.json文件
	if !isFile(relativePath + "/phpbear.json") {
		var bearOption = PhpBearOptions{
			FrameworkGit: FrameworkGit,
			FrameworkTag: FrameworkTag,
			ProjectGit:ProjectGit,
			InitTime: time.Now().Unix(),
		}

		data, err := json.Marshal(bearOption)
		if err != nil {
			return fmt.Errorf("构建phpbear.json失败: %v", err)
		}

		err = ioutil.WriteFile(relativePath + "/phpbear.json", data, 0644)
		if err != nil {
			return fmt.Errorf("构建phpbear.json失败：%v", err)
		}
	}

	return nil
}