//
// 3nigm4 ishtmdispatcher package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 06/10/2016
//

package main

// Arguments management struct.
type args struct {
	// server basic args
	verbose bool
	colored bool
	// mongodb
	dbAddresses string
	dbUsername  string
	dbPassword  string
	dbAuth      string
	// mail sender service
	senderAddress      string
	senderPort         int
	senderAuthUser     string
	senderAuthPassword string
	htmlTemplatePath   string
	// schedulers
	processScheduleMinutes  uint32
	dispatchScheduleMinutes uint32
	cleanupScheduleMinutes  uint32
}
