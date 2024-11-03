package app

import (
	"fmt"
	"html/template"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/gaowei-space/markdown-blog/internal/api"
	"github.com/gaowei-space/markdown-blog/internal/bindata/assets"
	"github.com/gaowei-space/markdown-blog/internal/bindata/views"
	"github.com/gaowei-space/markdown-blog/internal/types"
	"github.com/gaowei-space/markdown-blog/internal/utils"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/accesslog"
	"github.com/kataras/iris/v12/view"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
	"github.com/urfave/cli/v2"
)

var (
	MdDir      string
	Env        string
	Title      string
	Index      string
	ICP        string
	ISF        string
	FDir       string
	Copyright  int64
	LayoutFile = "layouts/layout.html"
	LogsDir    = "cache/logs/"
	TocPrefix  = "[toc]"
	IgnoreFile = []string{`favicon.ico`, `.DS_Store`, `.gitignore`, `README.md`}
	IgnorePath = []string{`.git`}
	Cache      time.Duration
	Analyzer   types.Analyzer
	Gitalk     types.Gitalk
)

// web服务器默认端口
const DefaultPort = 5006

func RunWeb(ctx *cli.Context) error {
	initParams(ctx)

	app := iris.New()

	setLog(app)

	app.RegisterView(getTmpl())
	app.OnErrorCode(iris.StatusNotFound, api.NotFound)
	app.OnErrorCode(iris.StatusInternalServerError, api.InternalServerError)

	setIndexAuto := false
	if Index == "" {
		setIndexAuto = true
	}

	app.Use(func(ctx iris.Context) {
		activeNav := getActiveNav(ctx)

		navs, firstNav := getNavs(activeNav)

		firstLink := utils.CustomURLEncode(strings.TrimPrefix(firstNav.Link, "/"))
		if setIndexAuto && Index != firstLink {
			Index = firstLink
		}

		// 设置 Gitalk ID
		Gitalk.Id = utils.MD5(activeNav)

		ctx.ViewData("Gitalk", Gitalk)
		ctx.ViewData("Analyzer", Analyzer)
		ctx.ViewData("Title", Title)
		ctx.ViewData("Nav", navs)
		ctx.ViewData("ICP", ICP)
		ctx.ViewData("ISF", ISF)
		ctx.ViewData("Copyright", Copyright)
		ctx.ViewData("ActiveNav", activeNav)
		ctx.ViewLayout(LayoutFile)

		ctx.Next()
	})

	app.Favicon("./favicon.ico")
	app.HandleDir("/static", getStatic())
	app.Get("/{f:path}", iris.Cache(Cache), articleHandler)
	app.Get(fmt.Sprintf("/%s/{f:path}", FDir), serveFileHandler)

	app.Run(iris.Addr(":" + strconv.Itoa(parsePort(ctx))))

	return nil
}

func getStatic() interface{} {
	if Env == "prod" {
		return assets.AssetFile()
	} else {
		return "./web/assets"
	}
}

func getTmpl() *view.HTMLEngine {
	if Env == "prod" {
		return iris.HTML(views.AssetFile(), ".html").Reload(true)
	} else {
		return iris.HTML("./web/views", ".html").Reload(true)
	}
}

func initParams(ctx *cli.Context) {
	MdDir = ctx.String("dir")
	if strings.TrimSpace(MdDir) == "" {
		log.Panic("Markdown files folder cannot be empty")
	}
	MdDir, _ = filepath.Abs(MdDir)

	Env = ctx.String("env")
	Title = ctx.String("title")
	Index = ctx.String("index")
	ICP = ctx.String("icp")
	ISF = ctx.String("isf")
	Copyright = ctx.Int64("copyright")
	FDir = ctx.String("fdir")

	Cache = time.Minute * 0
	if Env == "prod" {
		Cache = time.Minute * time.Duration(ctx.Int64("cache"))
	}

	// 设置分析器
	Analyzer.SetAnalyzer(ctx.String("analyzer-baidu"), ctx.String("analyzer-google"))

	// 设置Gitalk
	Gitalk.SetGitalk(ctx.String("gitalk.client-id"), ctx.String("gitalk.client-secret"), ctx.String("gitalk.repo"), ctx.String("gitalk.owner"), ctx.StringSlice("gitalk.admin"), ctx.StringSlice("gitalk.labels"))

	// 忽略文件
	IgnoreFile = append(IgnoreFile, ctx.StringSlice("ignore-file")...)
	IgnorePath = append(IgnorePath, FDir)
	IgnorePath = append(IgnorePath, ctx.StringSlice("ignore-path")...)
}

func setLog(app *iris.Application) {
	os.MkdirAll(LogsDir, 0777)
	f, _ := os.OpenFile(LogsDir+"access-"+time.Now().Format("20060102")+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)

	if Env == "prod" {
		app.Logger().SetOutput(f)
	} else {
		app.Logger().SetLevel("debug")
		app.Logger().Debugf(`Log level set to "debug"`)
	}

	// Close the file on shutdown.
	app.ConfigureHost(func(su *iris.Supervisor) {
		su.RegisterOnShutdown(func() {
			f.Close()
		})
	})

	ac := accesslog.New(f)
	ac.AddOutput(app.Logger().Printer)
	app.UseRouter(ac.Handler)
	app.Logger().Debugf("Using <%s> to log requests", f.Name())
}

func parsePort(ctx *cli.Context) int {
	port := DefaultPort
	if ctx.IsSet("port") {
		port = ctx.Int("port")
	}
	if port <= 0 || port >= 65535 {
		port = DefaultPort
	}

	return port
}

func getNavs(activeNav string) ([]map[string]interface{}, utils.Node) {
	var option utils.Option
	option.RootPath = []string{MdDir}
	option.SubFlag = true
	option.IgnorePath = IgnorePath
	option.IgnoreFile = IgnoreFile
	tree, _ := utils.Explorer(option)

	navs := make([]map[string]interface{}, 0)
	for _, v := range tree.Children {
		//进行排序
		utils.SortLess(v)
		for _, item := range v.Children {
			log.Printf("getNavs, %v, %v", item.Link, item.IsDir)
			searchActiveNav(item, activeNav)
			navs = append(navs, structs.Map(item))
		}
	}

	firstNav := getFirstNav(*tree.Children[0])

	return navs, firstNav
}

func searchActiveNav(node *utils.Node, activeNav string) {
	link_str, _ := url.QueryUnescape(node.Link)
	if !node.IsDir && strings.TrimPrefix(link_str, "/") == activeNav {
		node.Active = "active"
		return
	}
	if len(node.Children) > 0 {
		for _, v := range node.Children {
			searchActiveNav(v, activeNav)
		}
	}
}

func getFirstNav(node utils.Node) utils.Node {
	if !node.IsDir {
		return node
	}
	return getFirstNav(*node.Children[0])
}

func getActiveNav(ctx iris.Context) string {
	f := ctx.Params().Get("f")
	if f == "" {
		f = Index
	}
	return f
}

func serveFileHandler(ctx iris.Context) {
	f := ctx.Params().Get("f")
	file := MdDir + "/" + FDir + "/" + f
	ctx.ServeFile(file)
}

func articleHandler(ctx iris.Context) {
	f := getActiveNav(ctx)

	if utils.IsInSlice(IgnoreFile, f) {
		return
	}

	// mdfile := MdDir + "/" + f + ".md"
	mdfile := MdDir + "/" + f

	_, err := os.Stat(mdfile)
	if err != nil {
		ctx.StatusCode(404)
		ctx.Application().Logger().Errorf("Not Found '%s', Path is %s", mdfile, ctx.Path())
		return
	}

	bytes, err := os.ReadFile(mdfile)
	if err != nil {
		ctx.StatusCode(500)
		ctx.Application().Logger().Errorf("ReadFile Error '%s', Path is %s", mdfile, ctx.Path())
		return
	}
	tmp := strings.Split(f, "/")
	title := tmp[len(tmp)-1]
	ctx.ViewData("Title", title+" - "+Title)

	var ext = filepath.Ext(f)
	ctx.ViewData("Article", mdToHtml(bytes, ext))

	ctx.View("index.html")
}

func mdToHtml(content []byte, ext string) template.HTML {
	str := string(content)

	if ext != ".md" {
		log.Printf("mdToHtml, %v", ext)
		str = "```" + ext + "\n" + str + "\n```"
	}

	return mdToHtml2(str)
}

func mdToHtml2(strs string) template.HTML {

	var htmlFlags blackfriday.HTMLFlags

	if strings.HasPrefix(strs, TocPrefix) {
		htmlFlags |= blackfriday.TOC
		strs = strings.Replace(strs, TocPrefix, "<br/><br/>", 1)
	}

	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: htmlFlags,
	})

	// fix windows \r\n
	unix := strings.ReplaceAll(strs, "\r\n", "\n")

	unsafe := blackfriday.Run([]byte(unix), blackfriday.WithRenderer(renderer), blackfriday.WithExtensions(blackfriday.CommonExtensions))

	// 创建bluemonday策略，只允许<span>标签及其style属性
	p := bluemonday.UGCPolicy()
	p.AllowElements("span")                  // 只允许<span>标签
	p.AllowAttrs("style").OnElements("span") // 在<span>上允许使用style属性

	// 使用自定义的bluemonday策略来清理HTML
	html := p.SanitizeBytes(unsafe)

	return template.HTML(string(html))
}
