package main

import(
	"log"
	"os"
	"time"
	"os/signal"
	"syscall"
	"strings"
	"bufio"
)
type system struct {
	SystemName	string
	Mandant		string
	ServerIP	string
	Port		string
	ServicePath	string
	FetchSet	string
	AttachmentSet	string
	ActivateBCC	bool
	BasicAuthUser	string
	BasicAuthPW	string
}
func main() {
	log.Println("Go-App started")
	//graceful shutdown
	osCall := make(chan os.Signal, 1)
	signal.Notify(osCall, os.Interrupt, syscall.SIGTERM)
	//Time
	ticker := time.NewTicker(60 * time.Second)

	mc := NewMailCollector()
	go mc.Start(8080)
	systems := readConfig()
	doSystems(mc, systems)
	for {
		select {
		case <- osCall:
			log.Fatalln("Shutdown received")
			break
		case <- ticker.C:
			doSystems(mc, systems)
		}
	}

}
func doSystems(mc *MailCollector, systems []system){
	log.Println("[MAIN] Checking for Mails: Start")
	for _, s := range systems{
		go doSystem(s, mc)
	}
	log.Println("[MAIN] Checking for Mails: Finished")

}
func readConfig() []system {
	log.Println("[CFG] Reading Config File /app/data.csv")
	f, err := os.Open("/app/data.csv")
	if err != nil {
		log.Fatalln(err)
	}
	defer func(){
		log.Println("[CFG] Closing /app/data.csv")
		err = f.Close()
		if err != nil{
			log.Fatalln(err)
		}
	}()
	ret := []system{}
	s := bufio.NewScanner(f)
	for s.Scan(){
		if (strings.HasPrefix(s.Text(), "#")){
			log.Println("[CFG] Ignoring commented Line (starting with #")
			continue
		}
		log.Println("[CFG] Found new System!")
		newSystem := system{}
		for k, v := range strings.Split(s.Text(), ";"){
			switch(k){
			case 0:
				log.Printf("[CFG] Found Token \"SystemName\": \"%s\"\n", v)
				newSystem.SystemName = v
			case 1:
				log.Printf("[CFG] Found Token \"Mandant\": \"%s\"\n", v)
				newSystem.Mandant = v
			case 2:
				log.Printf("[CFG] Found Token \"ServerIP\": \"%s\"\n", v)
				newSystem.ServerIP = v
			case 3:
				log.Printf("[CFG] Found Token \"Port\": \"%s\"\n", v)
				newSystem.Port = v
			case 4:
				log.Printf("[CFG] Found Token \"ServicePath\": \"%s\"\n", v)
				newSystem.ServicePath = v
			case 5:
				log.Printf("[CFG] Found Token \"FetchSet\": \"%s\"\n", v)
				newSystem.FetchSet = v
			case 6:
				log.Printf("[CFG] Found Token \"AttachmentSet\": \"%s\"\n", v)
				newSystem.AttachmentSet = v
			case 7:
				log.Printf("[CFG] Found Token \"ActivateBCC\": \"%s\"\n", v)
				if (v == "true" || v == "yes") {
					newSystem.ActivateBCC = true
				}else{
					newSystem.ActivateBCC = false
				}
			case 8:
				log.Printf("[CFG] Found Token \"BasicAuthUser\": \"%s\"\n", v)
				newSystem.BasicAuthUser = v
			case 9:
				log.Printf("[CFG] Found Token \"BasicAuthPW\", not showing it in logs\n")
				newSystem.BasicAuthPW = v
			default:
				log.Printf("[CFG] Unknown Token \"%s\" at index %d\n", v, k)
			}
		}
		ret = append(ret, newSystem)

	}

	log.Printf("[CFG] Found %d parsable lines in /app/data.csv\n", len(ret))
	return ret

}
