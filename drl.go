package drl

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Server struct {
	HostName   string
	ID         string
	LoadPerSec int64
	Percentage float64
}

type DRL struct {
	Servers           *Cache
	mutex             sync.Mutex
	serverIndex       map[string]bool
	ThisServerID      string
	CurrentTotal      int64
	RequestTokenValue int
	CurrentTokenValue int
}

func (d *DRL) Init() {
	d.Servers = NewCache(10 * time.Second)
	d.RequestTokenValue = 100
	d.mutex = sync.Mutex{}
	d.serverIndex = make(map[string]bool)
}

func (d *DRL) uniqueID(s Server) string {
	uniqueID := s.ID + "|" + s.HostName
	return uniqueID
}

func (d *DRL) totalLoadAcrossServers() int64 {
	var total int64
	toRemove := map[string]bool{}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	for sID, _ := range d.serverIndex {
		thisServerObject, found := d.Servers.Get(sID)
		if !found {
			toRemove[sID] = true
		} else {
			total += thisServerObject.LoadPerSec
		}
	}

	d.CurrentTotal = total

	// Update the server list
	for sID, _ := range toRemove {
		fmt.Println("Removing: ")
		fmt.Println(sID)
		delete(d.serverIndex, sID)
	}

	return total
}

func (d *DRL) percentagesAcrossServers() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	for sID, _ := range d.serverIndex {
		thisServerObject, found := d.Servers.Get(sID)
		if found {
			curTot := d.CurrentTotal
			if d.CurrentTotal == 0 {
				curTot = 1
			}
			thisServerObject.Percentage = float64(thisServerObject.LoadPerSec) / float64(curTot)
			d.Servers.Set(sID, thisServerObject)
		}
	}
}

func (d *DRL) calculateTokenBucketValue() error {
	thisServerObject, found := d.Servers.Get(d.ThisServerID)
	if !found {
		return errors.New("Apprently this server does not exist!")
	}

	var thisTokenValue float64
	thisTokenValue = float64(d.RequestTokenValue)

	if thisServerObject.Percentage > 0 {
		thisTokenValue = float64(d.RequestTokenValue) / thisServerObject.Percentage
	}

	rounded := Round(thisTokenValue, .5, 0)
	d.CurrentTokenValue = int(rounded)
	return nil
}

func (d *DRL) AddOrUpdateServer(s Server) error {
	// Add or update the cache
	d.mutex.Lock()
	d.serverIndex[d.uniqueID(s)] = true
	d.mutex.Unlock()

	d.Servers.Set(d.uniqueID(s), s)

	// Recalculate totals
	d.totalLoadAcrossServers()

	// Recalculate percentages
	d.percentagesAcrossServers()

	// Get the current token bucket value:
	calcErr := d.calculateTokenBucketValue()
	if calcErr != nil {
		return calcErr
	}

	return nil
}
