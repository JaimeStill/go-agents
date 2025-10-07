package capabilities_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/JaimeStill/go-agents/pkg/capabilities"
	"github.com/JaimeStill/go-agents/pkg/protocols"
)

func TestRegisterFormat_AndGetFormat(t *testing.T) {
	formatName := "test-format-1"

	capabilities.RegisterFormat(formatName, func() capabilities.Capability {
		return capabilities.NewChatCapability(formatName, nil)
	})

	cap, err := capabilities.GetFormat(formatName)
	if err != nil {
		t.Fatalf("GetFormat failed: %v", err)
	}

	if cap.Name() != formatName {
		t.Errorf("got name %q, want %q", cap.Name(), formatName)
	}

	if cap.Protocol() != protocols.Chat {
		t.Errorf("got protocol %q, want %q", cap.Protocol(), protocols.Chat)
	}
}

func TestGetFormat_NonExistent(t *testing.T) {
	_, err := capabilities.GetFormat("non-existent-format")
	if err == nil {
		t.Error("expected error for non-existent format, got nil")
	}
}

func TestListFormats_Multiple(t *testing.T) {
	// Register multiple test formats
	testFormats := []string{
		"test-list-1",
		"test-list-2",
		"test-list-3",
	}

	for _, name := range testFormats {
		formatName := name
		capabilities.RegisterFormat(formatName, func() capabilities.Capability {
			return capabilities.NewChatCapability(formatName, nil)
		})
	}

	formats := capabilities.ListFormats()

	// Check that our test formats are included
	foundCount := 0
	for _, format := range formats {
		for _, testFormat := range testFormats {
			if format == testFormat {
				foundCount++
				break
			}
		}
	}

	if foundCount != len(testFormats) {
		t.Errorf("found %d test formats in list, want %d", foundCount, len(testFormats))
	}
}

func TestRegistry_ConcurrentRegistration(t *testing.T) {
	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(index int) {
			defer wg.Done()

			formatName := fmt.Sprintf("concurrent-test-%d", index)
			capabilities.RegisterFormat(formatName, func() capabilities.Capability {
				return capabilities.NewChatCapability(formatName, nil)
			})
		}(i)
	}

	wg.Wait()

	// Verify all formats were registered
	formats := capabilities.ListFormats()

	foundCount := 0
	for i := 0; i < goroutines; i++ {
		expectedName := fmt.Sprintf("concurrent-test-%d", i)
		for _, format := range formats {
			if format == expectedName {
				foundCount++
				break
			}
		}
	}

	if foundCount != goroutines {
		t.Errorf("found %d concurrent formats, want %d", foundCount, goroutines)
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	// Register a format for concurrent access testing
	formatName := "concurrent-access-test"
	capabilities.RegisterFormat(formatName, func() capabilities.Capability {
		return capabilities.NewChatCapability(formatName, nil)
	})

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			_, err := capabilities.GetFormat(formatName)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	errorCount := 0
	for err := range errors {
		t.Errorf("concurrent access error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Fatalf("had %d errors during concurrent access", errorCount)
	}
}
