//
// 3nigm4 ishtm package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 11/09/2016
//

package ishtm

// Golang std packages
import (
	"testing"
	"time"
)

func TestNewJob(t *testing.T) {
	owner := &OwnerID{
		Name:        "userA",
		Email:       "userA@mail.com",
		KeyID:       23984,
		Fingerprint: []byte("a3ir0ffe30b1fa2"),
	}
	settings := &Settings{
		ExtensionUnit:  time.Duration(3000 * time.Millisecond),
		DisableOffset:  true,
		NotifyDeadline: true,
		DeliveryOffset: time.Duration(3000 * time.Millisecond),
	}
	recipients := make([]Recipient, 0)

	GlobalEncryptionKey = []byte("thisisatesttempkeyiroeofod090877")
	GlobalEncryptionSalt = []byte("thisissa")

	now := time.Now().UTC()
	reference := now.Add(settings.DeliveryOffset)
	job, _, err := NewJob(owner, []byte("This is a mock reference file"), settings, recipients)
	if err != nil {
		t.Fatalf("Unable to create job instance: %s.\n", err.Error())
	}
	if now.Sub(job.Creation) > 1*time.Millisecond {
		t.Fatalf("Diff %d on creation is more than tolerated delta.\n", now.Sub(job.Creation))
	}
	if now.Sub(job.LastModified) > 1*time.Millisecond {
		t.Fatalf("Diff %d on last mod is more than tolerated delta.\n", now.Sub(job.LastModified))
	}
	if now.Sub(job.LastPing) > 1*time.Millisecond {
		t.Fatalf("Diff %d on last ping is more than tolerated delta.\n", now.Sub(job.LastPing))
	}

	if job.TimeToDelivery.Sub(reference)-(3000*time.Millisecond) > 1*time.Millisecond {
		t.Fatalf("Unexpected ttd delta %d expecting %d.\n", job.TimeToDelivery.Sub(reference), 3000*time.Millisecond)
	}
}

func TestRefresTtd(t *testing.T) {
	owner := &OwnerID{
		Name:        "userA",
		Email:       "userA@mail.com",
		KeyID:       23984,
		Fingerprint: []byte("a3ir0ffe30b1fa2"),
	}
	settings := &Settings{
		ExtensionUnit:  time.Duration(3000 * time.Millisecond),
		DisableOffset:  true,
		NotifyDeadline: true,
		DeliveryOffset: time.Duration(3000 * time.Millisecond),
	}
	recipients := make([]Recipient, 0)

	GlobalEncryptionKey = []byte("thisisatesttempkeyiroeofod090877")
	GlobalEncryptionSalt = []byte("thisissa")

	now := time.Now().UTC()
	reference := now.Add(settings.DeliveryOffset)
	job, _, err := NewJob(owner, []byte("This is a mock reference file"), settings, recipients)
	if err != nil {
		t.Fatalf("Unable to create job instance: %s.\n", err.Error())
	}
	if reference.Unix()-job.TimeToDelivery.Unix() != 0 {
		t.Fatalf("Expected ttd time should be %v but found %v.\n",
			now.Add(time.Duration(3000*time.Millisecond)),
			job.TimeToDelivery,
		)
	}
	time.Sleep(500 * time.Microsecond)

	err = job.Refresh()
	if err != nil {
		t.Fatalf("Error while refreshing: %s.\n", err.Error())
	}
	// max tolerance 1 ms
	if job.TimeToDelivery.Sub(reference)-(3000*time.Millisecond) > 1*time.Millisecond {
		t.Fatalf("Unexpected ttd delta %d expecting %d.\n", job.TimeToDelivery.Sub(reference), 3000*time.Millisecond)
	}
	newReference := job.TimeToDelivery

	err = job.Refresh()
	if err != nil {
		t.Fatalf("Error while refreshing: %s.\n", err.Error())
	}
	// max tolerance 1 ms
	if job.TimeToDelivery.Sub(newReference)-(3000*time.Millisecond) > 1*time.Millisecond {
		t.Fatalf("Unexpected ttd delta %d expecting %d.\n", job.TimeToDelivery.Sub(newReference), 3000*time.Millisecond)
	}
}
