package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)


func watchConfig(configYAML *map[string]interface{}, configFilePath *string){
	watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Close()
	go func() {
        for {
            select {
            case event, ok := <-watcher.Events:
                if !ok {
                    return
                }
                if event.Has(fsnotify.Write) {
					yamlFile, err := os.ReadFile(*configFilePath)
					if err != nil {
						fmt.Printf("yamlFile.Get err #%v ", err)
					}
					err = yaml.Unmarshal(yamlFile, &configYAML)
					if err != nil {
						fmt.Printf("Unmarshal: %v", err)
					}
                }
            case err, ok := <-watcher.Errors:
                if !ok {
                    return
                }
                log.Println("error:", err)
            }
        }
    }()
	    // Add a path.
		err = watcher.Add("config.yaml")
		if err != nil {
			log.Fatal(err)
		}
	
		// Block main goroutine forever.
		<-make(chan struct{})
}

func getConfig(path string) (map[string]interface{}){


	configYAML := make(map[string]interface{})

	yamlFile, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("yamlFile.Get err #%v ", err)
		os.Exit(1)
	}
	err = yaml.Unmarshal(yamlFile, configYAML)
	if err != nil {
		fmt.Printf("Unmarshal: %v", err)
		os.Exit(1)
	}

	
	
	return configYAML
}


func main() {
	configFilePath := "config.yaml"
	configYAML := getConfig(configFilePath)
	langFile, _ := configYAML["lang_file"].(string)
	var langYAML map[string]interface{}
	if _, ok := configYAML["lang_file"].(string); ok {
		langFilePath := "langs/"+langFile
		langYAML = getConfig(langFilePath)
	}



	// Flags
	var watchFlag bool
	var debugFlag bool
	flag.BoolVar(&watchFlag, "w", false, "Watch config file for changes")
	flag.BoolVar(&debugFlag, "d", false, "Debug mode")
	flag.Parse()

	if watchFlag{
		go watchConfig(&configYAML, &configFilePath)
	}

	if !debugFlag {
		gin.SetMode(gin.ReleaseMode)
	}
	// End Flags




	router := gin.Default()

	// Static Router
	router.Static("/static", "static")
	// End Static Router

	// Custom Pages Router
	customPages := configYAML["custom_pages"].([]interface {})
	mapString := make(map[string]string)
	for _, value := range customPages {
		for key, test := range value.(map[interface {}]interface {}){
			strKey := fmt.Sprintf("%v", key)
			strValue := fmt.Sprintf("%v", test)
			mapString[strKey] = strValue
		}
		routePath := "/" + mapString["name"]
		router.GET(routePath, func(c *gin.Context) {
			c.HTML(http.StatusOK, mapString["file"], gin.H{
				"uri":    "http://" + c.Request.Host,
				"config": configYAML,
				"lang": langYAML,
				"customPages" : customPages,
			})
		})
	}
	// End Custom Pages Router

	// Index Router
	router.LoadHTMLGlob("templates/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"uri":    "http://" + c.Request.Host,
			"config": configYAML,
			"lang": langYAML,
			"customPages" : customPages,
		})
	})
   // End Index Router


	if debugFlag {
		router.Run(":8090")
	} else{
		router.Run("0.0.0.0:80")
	}
	
}
