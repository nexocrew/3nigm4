//
// 3nigm4 will package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 11/09/2016
//

package will

// Golang std pkgs
import (
	"testing"
	"time"
)

// Internal pkgs
import (
	ct "github.com/nexocrew/3nigm4/lib/commons"
)

func TestNewWill(t *testing.T) {
	owner := &OwnerID{
		Name:  "userA",
		Email: "userA@mail.com",
	}
	settings := &Settings{
		ExtensionUnit:  time.Duration(3000 * time.Millisecond),
		DisableOffset:  true,
		NotifyDeadline: true,
		DeliveryOffset: time.Duration(3000 * time.Millisecond),
	}
	recipients := make([]ct.Recipient, 0)

	GlobalEncryptionKey = []byte("thisisatesttempkeyiroeofod090877")
	GlobalEncryptionSalt = []byte("thisissa")

	now := time.Now().UTC()
	reference := now.Add(settings.DeliveryOffset)
	will, credentials, err := NewWill(owner, []byte("This is a mock reference file"), settings, recipients)
	if err != nil {
		t.Fatalf("Unable to create will instance: %s.\n", err.Error())
	}
	if now.Sub(will.Creation) > 1*time.Millisecond {
		t.Fatalf("Diff %d on creation is more than tolerated delta.\n", now.Sub(will.Creation))
	}
	if now.Sub(will.LastModified) > 1*time.Millisecond {
		t.Fatalf("Diff %d on last mod is more than tolerated delta.\n", now.Sub(will.LastModified))
	}
	if now.Sub(will.LastPing) > 1*time.Millisecond {
		t.Fatalf("Diff %d on last ping is more than tolerated delta.\n", now.Sub(will.LastPing))
	}

	if will.TimeToDelivery.Sub(reference)-(3000*time.Millisecond) > 1*time.Millisecond {
		t.Fatalf("Unexpected ttd delta %d expecting %d.\n", will.TimeToDelivery.Sub(reference), 3000*time.Millisecond)
	}

	if len(credentials.SecondaryKeys) != secondaryKeysNumber {
		t.Fatalf("Should have %d secondary keys at generation but found %d.\n", secondaryKeysNumber, len(credentials.SecondaryKeys))
	}
	for idx, key := range credentials.SecondaryKeys {
		if len(key) == 0 {
			t.Fatalf("Secondary keys must not be null.\n")
		}
		t.Logf("Secondary key %d: %s", idx, key)
	}
}

func TestRefresTtd(t *testing.T) {
	owner := &OwnerID{
		Name:  "userA",
		Email: "userA@mail.com",
	}
	settings := &Settings{
		ExtensionUnit:  time.Duration(3000 * time.Millisecond),
		DisableOffset:  true,
		NotifyDeadline: true,
		DeliveryOffset: time.Duration(3000 * time.Millisecond),
	}
	recipients := make([]ct.Recipient, 0)

	GlobalEncryptionKey = []byte("thisisatesttempkeyiroeofod090877")
	GlobalEncryptionSalt = []byte("thisissa")

	now := time.Now().UTC()
	reference := now.Add(settings.DeliveryOffset)
	will, _, err := NewWill(owner, []byte("This is a mock reference file"), settings, recipients)
	if err != nil {
		t.Fatalf("Unable to create will instance: %s.\n", err.Error())
	}
	if reference.Unix()-will.TimeToDelivery.Unix() != 0 {
		t.Fatalf("Expected ttd time should be %v but found %v.\n",
			now.Add(time.Duration(3000*time.Millisecond)),
			will.TimeToDelivery,
		)
	}
	time.Sleep(500 * time.Microsecond)

	err = will.Refresh()
	if err != nil {
		t.Fatalf("Error while refreshing: %s.\n", err.Error())
	}
	// max tolerance 1 ms
	if will.TimeToDelivery.Sub(reference)-(3000*time.Millisecond) > 1*time.Millisecond {
		t.Fatalf("Unexpected ttd delta %d expecting %d.\n", will.TimeToDelivery.Sub(reference), 3000*time.Millisecond)
	}
	newReference := will.TimeToDelivery

	err = will.Refresh()
	if err != nil {
		t.Fatalf("Error while refreshing: %s.\n", err.Error())
	}
	// max tolerance 1 ms
	if will.TimeToDelivery.Sub(newReference)-(3000*time.Millisecond) > 1*time.Millisecond {
		t.Fatalf("Unexpected ttd delta %d expecting %d.\n", will.TimeToDelivery.Sub(newReference), 3000*time.Millisecond)
	}
}
