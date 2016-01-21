package main

import (
    "os"
    "fmt"
    "net"
    "net/http"
    "net/url"
    "crypto/tls"
    "io/ioutil"
    "bytes"
    "strings"
    "encoding/json"
    
    flags "github.com/jessevdk/go-flags"
    log "github.com/Sirupsen/logrus"
    
    "database/sql"
    "github.com/go-sql-driver/mysql"
)

var version string = "undef"

type Options struct {
    ShowHelp bool      `long:"help"`
    Debug bool     `env:"DEBUG"    long:"debug"    description:"enable debug"`
    
    ExternalIPLookupURL string `long:"external-ip-lookup" default:"http://ifconfig.co/"`
    
    // mysql options that we're emulating
    Host      string `short:"h" long:"host"                   required:"true" description:"Connect to host"`
    Port      uint32 `short:"P" long:"Port"                   default:"3306"  description:"Port number to use for connection"`
    User      string `short:"u" long:"user"                   required:"true" description:"User for login"`
    Password  string `short:"p" long:"password"               required:"true" description:"Password to use when connecting to server"`
    SSLCA     string `          long:"ssl-ca"                                 description:"CA file in PEM format"`
    SSLVerify bool   `          long:"ssl-verify-server-cert" default:"false" description:"Verify server's \"Common Name\" in its cert against hostname used when connecting."`
    
    Args struct {
        DBName string
    } `positional-args:"true" required:"true"`
}

type Result struct {
    Version string
    Options Options
    DBName  string
    DSN     string

    ExternalIP string
    MySQLServerAddr *net.IPAddr
    
    Passed bool
    Result string
}

func main() {
    var opts Options
    
    flagsParser := flags.NewParser(&opts, flags.PrintErrors | flags.PassDoubleDash)
    _, err := flagsParser.Parse()
    if err != nil {
        if opts.ShowHelp {
            var b bytes.Buffer

            flagsParser.WriteHelp(&b)
            print(b.String())
        }
        
        os.Exit(1)
    }
    
    if opts.Debug {
        log.SetLevel(log.DebugLevel)
    }
    
    result := &Result{
        Version: version,
        Options: opts,
        DBName: opts.Args.DBName,
    }
    
    log.Debugf("looking up external IP")
    extIpReq, err := http.NewRequest("GET", opts.ExternalIPLookupURL, nil)
    checkError("creating http request", err)
    
    // these headers and their values are required by ifconfig.co ¯\_(ツ)_/¯ 
    extIpReq.Header.Set("Accept", "text/plain")
    extIpReq.Header.Set("User-Agent", "curl/1.2.3")
    extIpResp, err := http.DefaultClient.Do(extIpReq)
    checkError("looking up external IP", err)
    
    defer extIpResp.Body.Close()
    body, err := ioutil.ReadAll(extIpResp.Body)
    checkError("reading body of external IP request", err)

    result.ExternalIP = strings.TrimSpace(string(body))

    log.Debugf("resolving hostname")
    result.MySQLServerAddr, err = net.ResolveIPAddr("ip", opts.Host)
    checkError(fmt.Sprintf("error resolving %s", opts.Host), err)
    
    dsnParams := &url.Values{}
    dsnParams.Set("timeout", "30s")
    
    if opts.SSLCA != "" {
        mysql.RegisterTLSConfig("pre-resolved", &tls.Config{
            ServerName: opts.Host,
            InsecureSkipVerify: ! opts.SSLVerify,
        })

        dsnParams.Set("tls", "pre-resolved")
    }
    
    var resolvedAddr string
    if result.MySQLServerAddr.Zone != "" {
        // ipv6
        resolvedAddr = "[" + result.MySQLServerAddr.String() + "]"
    } else {
        resolvedAddr = result.MySQLServerAddr.String()
    }

    // DSN format:
    // [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
    result.DSN = fmt.Sprintf(
        "%s:%s@tcp(%s:%d)/%s?%s",
        opts.User,
        opts.Password,
        resolvedAddr,
        opts.Port,
        opts.Args.DBName,
        dsnParams.Encode(),
    )
    
    log.Debug("connecting to database")
    dbConn, err := sql.Open("mysql", result.DSN)
    checkError("unable to build connection", err)
    defer dbConn.Close()
    
    err = dbConn.Ping()
    if err != nil {
        result.Passed = false
        result.Result = fmt.Sprintf("error connecting: %v", err)
    } else {
        result.Passed = true
        result.Result = "successfully connected to database"
    }
    
    log.Debugf("result: %+v", result)
    
    jsonResult, err := json.MarshalIndent(result, "", "    ")
    checkError("marshalling json", err)
    println(string(jsonResult))
}
