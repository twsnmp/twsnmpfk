package datastore

import (
	"context"
	"os"
	"sync"
	"testing"
)

func TestNodeImage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	td, err := os.MkdirTemp("", "twsnmpfk_node_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)

	wg := &sync.WaitGroup{}
	Init(ctx, td, wg)
	defer CloseDB()

	dummyNode := &NodeEnt{
		Name: "Test Node",
		IP:   "192.168.1.10",
	}
	if err := AddNode(dummyNode); err != nil {
		t.Fatalf("AddNode failed: %v", err)
	}
	nodeID := dummyNode.ID

	// 1. Initial should be empty
	if img := GetNodeImage(nodeID); img != "" {
		t.Errorf("expected empty image, got %s", img)
	}

	// 2. Save image
	testImg := "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD..."
	if err := SaveNodeImage(nodeID, testImg); err != nil {
		t.Fatalf("SaveNodeImage failed: %v", err)
	}
	if img := GetNodeImage(nodeID); img != testImg {
		t.Errorf("expected %s, got %s", testImg, img)
	}

	// 3. Clear image by saving empty string
	if err := SaveNodeImage(nodeID, ""); err != nil {
		t.Fatalf("SaveNodeImage clear failed: %v", err)
	}
	if img := GetNodeImage(nodeID); img != "" {
		t.Errorf("expected empty image after clear, got %s", img)
	}

	// 4. Save image again, then delete node
	if err := SaveNodeImage(nodeID, testImg); err != nil {
		t.Fatalf("SaveNodeImage failed: %v", err)
	}
	if err := DeleteNode(nodeID); err != nil {
		t.Fatalf("DeleteNode failed: %v", err)
	}
	if img := GetNodeImage(nodeID); img != "" {
		t.Errorf("expected image to be deleted when node deleted, got %s", img)
	}
}
