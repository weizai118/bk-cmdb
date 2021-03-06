package command

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"configcenter/src/common"
	"configcenter/src/common/blog"
	"configcenter/src/common/conf"
	"configcenter/src/common/core/cc/api"
)

const bkbizCmdName = "bkbiz"

// Parse run app command
func Parse(args []string) error {
	if len(args) <= 1 || args[1] != bkbizCmdName {
		return nil
	}

	var (
		exportflag     bool
		importflag     bool
		miniflag       bool
		dryrunflag     bool
		filepath       string
		configposition string
		scope          string
	)

	// set flags
	bkbizfs := pflag.NewFlagSet(bkbizCmdName, pflag.ExitOnError)
	bkbizfs.BoolVar(&dryrunflag, "dryrun", false, "dryrun flag, if this flag seted, we will just print what we will do but not execute to db")
	bkbizfs.BoolVar(&exportflag, "export", false, "export flag")
	bkbizfs.BoolVar(&miniflag, "mini", false, "mini flag, only export required fields")
	bkbizfs.BoolVar(&importflag, "import", false, "import flag")
	bkbizfs.StringVar(&scope, "scope", "all", "export scope, could be [biz] or [process], default all")
	bkbizfs.StringVar(&filepath, "file", "", "export or import filepath")
	bkbizfs.StringVar(&configposition, "config", "conf/api.conf", "The config path. e.g conf/api.conf")
	err := bkbizfs.Parse(args[1:])
	if err != nil {
		return err
	}

	// init config
	config := new(conf.Config)
	config.InitConfig(configposition)

	// connect to mongo db
	a := api.NewAPIResource()
	err = a.GetDataCli(config.Configmap, "mongodb")
	if err != nil {
		blog.Error("connect mongodb error exit! err:%s", err.Error())
		return err
	}

	opt := &option{
		position: filepath,
		OwnerID:  common.BKDefaultOwnerID,
		dryrun:   dryrunflag,
		mini:     miniflag,
		scope:    scope,
	}

	if exportflag {
		mode := ""
		if miniflag {
			mode = "mini"
		} else {
			mode = "verbose"

		}
		fmt.Printf("exporting blueking business to %s in \033[34m%s\033[0m mode\n", filepath, mode)
		if err := export(a.InstCli, opt); err != nil {
			blog.Errorf("export error: %s", err.Error())
			os.Exit(2)
		}
		fmt.Printf("blueking business has been export to %s\n", filepath)
	} else if importflag {
		fmt.Printf("importing blueking business from %s\n", filepath)
		opt.mini = false
		opt.scope = "all"
		if err := importBKBiz(a.InstCli, opt); err != nil {
			blog.Errorf("import error: %s", err.Error())
			os.Exit(2)
		}
		if !dryrunflag {
			fmt.Printf("blueking business has been import from %s\n", filepath)
		}
	} else {
		blog.Errorf("invalide argument")
	}

	os.Exit(0)
	return nil
}
