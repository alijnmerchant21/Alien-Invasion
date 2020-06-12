package main

import (
        "strconv"
        "flag"
        "tenderfun/tendermint/util"
        "strings"
        "os"
        log "github.com/sirupsen/logrus"
        //"time"
)

// Assume that each Alien will start at a random city
// Assume that each Alien will move in a sequential fashion

// Dioikitís is the head of all the aliens, it control
// Dioikitís will only stop under three condition:
// 1) All the cities are destroyed
// 2) All the aliens are destroyed
// 3) All the iteration are exhausted (10000)


// TODO, change this back to 10000 once finished development
var numOfCommands = 10000

var dioikitísCounter = 1
var aliens map[Alien]Status
var cityLookup map[string]map[string]bool
var cityAlienLookup map[string]map[Alien]bool
var toCommanderSignalChan chan string
var teleiotísChan chan string
var gateKeeperChan chan bool

var syssoreftísSignalChan chan Alien
var teleiotísSumChan chan string
var teleiotísListDone  map[string]bool
var teleiotísListReady  map[string]bool

// to mark whether aliencommander had been finish or not
var alienCommanderComplete = false

func DioikitísCommander(iters int) <-chan int {
        c := make(chan int)
        go func() {
                for {
                        counter := 1
                        command, more := <-toCommanderSignalChan
                        if !alienCommanderComplete {
                                if more {
                                        if (command == "continue" && dioikitísCounter <= iters && len(aliens) > 0) {
                                                //log.Debugf("DioikitísCommander: I will trigger the %dth round of attack\n", dioikitísCounter)
                                                c <- counter
                                                dioikitísCounter++
                                                
                                        } else {
                                                //log.Debug("DioikitísCommander:  commander is exiting.....command", command, "counter: ", dioikitísCounter, "lens, ", len(aliens))
                                                close(c)
                                                alienCommanderComplete = true
                                        }
                                } else {
                                        //log.Debug("DioikitísCommander: All job are finished, commander is exiting.....")
                                        close(c)
                                        alienCommanderComplete = true
                                }
                        } else {
                                //log.Debug("DioikitísCommander: CommanderChan had been terminated alreayd!")
                        }
                }
        }()
        return c
}

func Move(alien Alien, lookupCity map[string]map[string]bool, cityAlienLookup map[string]map[Alien]bool) {
        oldCity := aliens[alien].city
        neighbour := lookupCity[oldCity]
        //log.Debug("Syssoreftís: what is the neighbour: ", neighbour, " for alien ", alien, ",", aliens[alien])
        if len(neighbour) == 0 {return}
        city := util.RandMove(neighbour)
        //log.Debug("Syssoreftís: what is the new city: ", city, " for alien ", aliens[alien])
        if city != "" {
                // remove previous alien from the old city from cityAlienlookup
                delete(cityAlienLookup[oldCity], alien)

                // move to the new city
                status := aliens[alien]
                status.city = city
                aliens[alien] = status

                // update cityAlienlookup
                aliensLookup := cityAlienLookup[city]
                if len(aliensLookup) == 0{
                        newMap := make(map[Alien]bool)
                        newMap[alien] = true

                        cityAlienLookup[city] = newMap
                } else{
                        aliensLookup[alien] = true
                }

        }
}

// After every move, check whether the city has two aliens stay there,
// if yes, then we need to clean up the lookupcity by update each city's
// neighbour and also update cityAlienlookup to remove the destroyed aliens
// and cities

func Check(alien Alien, lookupCity map[string]map[string]bool, cityAlienLookup map[string]map[Alien]bool) {
        city := aliens[alien].city
        alis := cityAlienLookup[city]
        //log.Debug("Syssoreftís: I'm in Check, aliens: ", aliens)
        //log.Debug("Syssoreftís: I'm in Check, cityAlienLookup",  cityAlienLookup)

        if (len(alis) > 1){
                ali1 := ""
                ali2 := ""

                // Clean up and destroy
                // send a termination signal to the aliens
                // update lookupcity
                // and citya connect to cityb,
                // then cityb must connect to citya
                indexer := 0
                for alien, _ := range alis {
                        close(alien.commandChan)
                        delete(alis, alien)
                        delete(aliens, alien)
                        if indexer == 0{
                                ali1 = alien.name
                                indexer++
                        } else {
                                ali2 = alien.name
                        }
                }

                log.Infof("Syssoreftís: %s has been destroyed by %s and %s!\n", city, ali1, ali2)

                delete(cityAlienLookup, city)

                for _, cmap := range lookupCity{
                        delete(cmap, city)
                }

                delete(lookupCity, city)
                teleiotísChan <- ali1 + "," + ali2
        } else {
                //log.Debug("Syssoreftís: Everything seems ok, sending info to teleiotísSumChan")
                teleiotísSumChan <- alien.name
        }
        //log.Debug("Syssoreftís: I'm in Check, lookupCity", lookupCity)
}

func consumer(alien Alien, syssoreftísSignalChan chan Alien) {
        for {
                _, more := <-alien.commandChan
                if more {
                        //log.Debug("Consumer: Alien ", alien.name, " had been activated and is in motion!")
                        syssoreftísSignalChan <- alien
                } else {
                        //log.Debug("Consumer: I'm done: ", alien.name)
                        teleiotísChan <- alien.name
                        return
                }
        }
}



// syssoreftís is in charge of collect the info from alien consumer and send the signal to aliencommander
func Syssoreftís() {
        go func() {
                //log.Debug("Syssoreftís, I'm in work...")
                for {
                        alien, more := <- syssoreftísSignalChan
                        if more{
                                //log.Debug("Syssoreftís: receive work...")

                                // make the move
                                // if able to move, then move and check the condition in the city
                                // if there is a hit (two aliens in the same city), then update the lookupcity and
                                // cityAlienlookup and destroy the alien
                                //log.Debug("Syssoreftís: I'm checking alien before the move: ", aliens[alien])
                                if _, ok := aliens[alien]; ok {
                                        Move(alien, cityLookup, cityAlienLookup)
                                        //log.Debug("Syssoreftís: I'm checking alien after the move: ", aliens[alien])

                                        status := aliens[alien]
                                        status.counter++
                                        aliens[alien] = status

                                        Check(alien, cityLookup, cityAlienLookup)
                                }
                        } else {
                                toCommanderSignalChan <- "done"
                                log.Infof("All the Aliens have been destroyed; Game Terminated!")
                        }
                }
        }()
}


func Teleiotís(size int){
        go func() {
                teleiotísListDone = make(map[string]bool)
                teleiotísListReady = make(map[string]bool)

                for {
                        //log.Debug("Teleiotís: Waiting for Aliens: ")

                        select {
                        case names, more := <-teleiotísChan :{
                                //log.Debug("Teleiotís: in teleiotísChan", names)
                                if more {
                                        for _, name := range strings.Split(names, ",") {
                                                teleiotísListDone[name] = true
                                                delete(teleiotísListReady, name)
                                        }

                                        //log.Debug("Teleiotís: The current status for jobs: ", len(teleiotísListDone), ",", len(teleiotísListReady))
                                        if len(teleiotísListDone) == size {
                                                close(syssoreftísSignalChan)
                                                gateKeeperChan <- true
                                                return
                                        } else if (len(teleiotísListDone) + len(teleiotísListReady) == size) {
                                                //log.Debug("Teleiotís: ALl aliens had made their move for the current round... Next round will start shortly...")
                                                teleiotísListReady = make(map[string]bool)
                                                toCommanderSignalChan <- "continue"
                                        }
                                } else {
                                        //log.Debug("Teleiotís: I'm done")
                                        gateKeeperChan <- true
                                        return
                                }
                        }
                        case name, _ := <-teleiotísSumChan :{
                                //log.Debug("Teleiotís: in teleiotísSumChan")

                                teleiotísListReady[name] = true
                                if (len(teleiotísListDone) + len(teleiotísListReady) == size) {
                                        //log.Debug("Teleiotís: ALl aliens had made their move for the current round. Next round will start shortly...")
                                        teleiotísListReady = make(map[string]bool)
                                        toCommanderSignalChan <- "continue"

                                }
                        }
                        }
                }
        }()
}

// This function is to fan out the command to the aliens from the commander
func Pass(ch <-chan int) {
        go func() {
                //log.Debug("Pass: I'm in work mode now")
                for i := range ch {
                        //log.Debug("Pass: what is the aliens now:", aliens)
                        for alien, _ := range aliens {
                                //log.Debug("Pass: sending to which alien:", alien)
                                alien.commandChan <- i
                        }
                }

                //log.Debug("Pass: I'm done here...")
                for alien, _ := range aliens {
                        // close all our fanOut channels when the input channel is exhausted.
                        //log.Debug("Pass: Work done now, close channels")
                        close(alien.commandChan)
                }
        }()

}

type Alien struct {
        name        string
        commandChan chan int
}

type Status struct{
        city string
        counter int
}

func GenerateAliens(num int, cities []string) map[Alien]Status {
        // use the int to keep track of which round the alien is at
        alis := make(map[Alien]Status)

        for ali := 1; ali <= num; ali++ {
                name := "Alien-" + strconv.Itoa(ali)
                city := util.RandCity(cities)

                // to fix the bug of n (n >1) aliens are born in
                // the same city and next able to clear out
                addAlien := true
                for alien, status := range alis {
                        if status.city == city{
                                log.Infof("%s has been destroyed by %s and %s!\n", city, name, alien.name)
                                delete(alis, alien)
                                addAlien = false

                                // remove from cities
                                for index, str := range cities {
                                        if str == city {
                                                cities[len(cities)-1], cities[index] = cities[index], cities[len(cities)-1]
                                                cities = cities[:len(cities)-1]
                                                break
                                        }
                                }

                                // remove from cityLookup(on both the key and value)
                                for _, cmap := range cityLookup{
                                        delete(cmap, city)
                                }
                                delete(cityLookup, city)
                                break
                        }
                }

                if addAlien {
                        ch := make(chan int)
                        alien := Alien{name, ch}
                        status := Status{city, 0}
                        // initialize with 0
                        alis[alien] = status
                }
        }

        //log.Debug("Factory, my aliens: ", alis)
        return alis
}

// generate a map that for a given city, ablt to check what aliens are in the city
func GenerateLookup(aliens map[Alien]Status) map[string]map[Alien]bool {
        result := make(map[string](map[Alien]bool), len(cityLookup))
        for city, _ := range cityLookup {
                for alien, status := range aliens {
                        if status.city == city {
                                m := make(map[Alien]bool)
                                m[alien] = true
                                result[city] = m
                        }
                }
        }
        return result
}



func initParas(numOfAliens int, mapFile string) {
        cityLookup = util.ParseCity(mapFile)
        cities := make([]string, len(cityLookup))

        // populate the cities
        cnt := 0
        for city, _ := range cityLookup {
                cities[cnt] = city
                cnt++
        }

        aliens = GenerateAliens(numOfAliens, cities)

        // given a city, which alien(s) is/are there
        cityAlienLookup = GenerateLookup(aliens)

        // To notify commander that downstream are clear and ready to make the next move
        toCommanderSignalChan = make(chan string)

        teleiotísChan = make(chan string)

        gateKeeperChan = make(chan bool)

        syssoreftísSignalChan = make(chan Alien)

        teleiotísSumChan = make(chan string)
}

func init() {
        // Output to stdout instead of the default stderr
        // Can be any io.Writer, see below for File example
        log.SetOutput(os.Stdout)
}


func main() {
        //log.Debug("Failed to log to file, using default stderr")
        var numOfAliens int
        var mapFile string
        var logLevel string

        flag.IntVar(&numOfAliens, "numofaliens", 2, "a integer that indicate the number of of aliens")
        flag.StringVar(&mapFile, "mapfile", "", "the file that contains the map and direction information")
        flag.StringVar(&logLevel, "loglevel", "info", "the level of the log for the program")

        flag.Parse()

        // Only log the warning severity or above.
        switch logLevel {
        case "info" :
                log.SetLevel(log.InfoLevel)

        case "warn":
                log.SetLevel(log.WarnLevel)

        case "debug":
                log.SetLevel(log.DebugLevel)

        case "error":
                log.SetLevel(log.ErrorLevel)

        case "fatal":
                log.SetLevel(log.FatalLevel)

        default:
                log.SetLevel(log.PanicLevel)

        }

        initParas(numOfAliens, mapFile)

        //log.Infoff("Hello there, there are total", len(cityLookup), " cities and ", numOfAliens, " aliens", ,numOfAliens)
        log.Infof("Hello!, there are total of %d Cities and %d Aliens", len(cityLookup), numOfAliens)

        // to initialize the job
        go func(){toCommanderSignalChan <- "continue"}()

        // start the commander:
        commandChannel := DioikitísCommander(numOfCommands)

        Pass(commandChannel)

        Syssoreftís()

        Teleiotís(len(aliens))

        for alien, _ := range aliens {
                go consumer(alien, syssoreftísSignalChan)
        }

        <-gateKeeperChan

        if (len(cityLookup) == 0) {
                log.Infof("There is no city left!")
        } else {
                str := ""
                for city, _ := range cityLookup {
                        str += city + " "
                }

                log.Infof("The city that  still exist is/are: ", str)
        }

        log.Infof("Cool, game finished, hope you enjoyed it!")
}