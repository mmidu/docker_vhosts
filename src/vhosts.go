package main

import (
	"fmt"
	"os"
	"bufio"
	"strings"
	"io/ioutil"
	"regexp"
	"runtime"
)

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func read() string {
	reader := bufio.NewReader(os.Stdin)
	str, err := reader.ReadString('\n')
	check(err)
	return strings.TrimRight(str, "\r\n")
}

func getHostsFile() (path string, status bool) {
	
	switch platform := runtime.GOOS; platform {
		case "darwin":
			status = true
			path =  "/private/etc/hosts"
		case "linux":
			status = true
			path =  "/etc/hosts"
		case "windows":
			status = true
			path = "c:\\Windows\\System32\\Drivers\\etc\\hosts"
		default:
			fmt.Printf("%s: your system is not yet supported. Concact for information.\n", platform)
			status = false
			path = ""
	}
	return

}

func writeHost(hostname string) {
	path, status := getHostsFile()

	if !status {
		os.Exit(1)
	}
	
	file, err := ioutil.ReadFile(path)
	check(err)
	
	match, _ := regexp.MatchString(hostname+`(\s|$)`, string(file))

	if match {
		fmt.Println("hostname already exists")
		os.Exit(1)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	check(err)
	f.WriteString("\n127.0.0.1\t"+hostname+"\n")
	f.Close()
}

func makeVhostsDir(confPath string) (string, string) {
	folderPath := ""
	for i := len(confPath)-1; i >= 0; i-- {
		if string(confPath[i]) == "/" || string(confPath[i]) == "\\" {
			folderPath = confPath[0:i]
			break
		}
		if i == 0 {
			fmt.Println(confPath+ ": invalid path format")
			os.Exit(1)
		}
	}

	_ = os.Mkdir(folderPath + "/docker_vhosts", os.ModeDir)
	return folderPath, folderPath + "/docker_vhosts"
}

func escape(s string) string {
	s = strings.Replace(s, "\\", "\\\\", -1)
	s = strings.Replace(s, "/", "\\/", -1)
	s = strings.Replace(s, ".", "\\.", -1)
	s = strings.Replace(s, "*", "\\*", -1)
	return s
}

func addVhost(hostname string, port string, vhostsDirPath string) {
	
	vhost := "<VirtualHost *:80 *:443>\n\tProxyPreserveHost On\n\tProxyRequests Off\n\tServerName "+hostname+"\n\tServerAlias "+hostname+"\n\t<Location \"/\">\n\t\tProxyPass http://127.0.0.1:"+port+"/\n\t\tProxyPassReverse http://"+hostname+"/\n\t</Location>\n</VirtualHost>"
	vhostPath := vhostsDirPath + "/" + regexp.MustCompile("\\.").ReplaceAllString(hostname,"-") + ".conf"
    f, err := os.OpenFile(vhostPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	check(err)
	f.WriteString(vhost)
	f.Close()
}

func add(){
	fmt.Print("Insert your apache configuration file path (might be named apache2.conf or httpd.conf): ")
	confPath := read()
	confPath = regexp.MustCompile("\\\\").ReplaceAllString(confPath,"/")

	_, vhostsDirPath := makeVhostsDir(confPath)
	
	file, err := ioutil.ReadFile(confPath)

	check(err)
	confline := escape("Include "+vhostsDirPath+"/*.conf")

	match, _ := regexp.MatchString(confline, string(file))

	if !match {
		f, err := os.OpenFile(confPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		check(err)
		f.WriteString("Include "+vhostsDirPath+"/*.conf\n")
		f.Close()
	}
	
	fmt.Print("hostname: ")
	hostname := read()

	writeHost(hostname)

	fmt.Print("port: ")
	port := read()

	addVhost(hostname, port, vhostsDirPath)
}

func remove(){
	fmt.Print("Insert your apache configuration file path (might be named apache2.conf or httpd.conf): ")
	confPath := read()
	
	confPath = regexp.MustCompile("\\\\").ReplaceAllString(confPath,"/")

	_, vhostsDirPath := makeVhostsDir(confPath)
	
	fmt.Print("hostname: ")
	hostname := read()
	vhostPath := vhostsDirPath + "/" + hostname + ".conf"

	if _, err := os.Stat(vhostPath); err == nil {
		os.Remove(vhostPath)
	} else {
		fmt.Println(hostname + " doesnt exists.")
	}

	path, status := getHostsFile()
	
	if status {
		file, err := ioutil.ReadFile(path)
		check(err)
		re := regexp.MustCompile("127.0.0.1\t"+hostname+"\n")
		res := re.ReplaceAllString(string(file), "")
		f, err := os.OpenFile(path, os.O_TRUNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		f.WriteString(res)
		f.Close()
	}
	
}

func promptAction(){
	fmt.Print("Do you want to add o remove a docker virtual host? (a/r) ")
	action := read()
	l_action := strings.ToLower(action)

	switch l_action {
		case "a", "add":
			add()
			break
		case "r", "remove":
			remove()
			break
		case "exit":
			return
		default:
			fmt.Println(action + " is not a valid action.")
			promptAction()
	}

}

func main() {
	fmt.Println("WARNING: in order to use proxies, you have to enable the modules proxy and proxy_http on your Apache.")
	promptAction()
}
