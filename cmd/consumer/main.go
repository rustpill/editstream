package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rustpill/editstream/internal/edit"
)

const topic = "edits"

func main() {
	// replay flag enable replay mode
	replay := flag.Bool("replay", false, "rebuild the view: consume from offset 0 under a fresh group id")
	flag.Parse()

	// every replay uses a brand new consumer group ID
	// no memory of previous offset
	group := "leaderboard"
	if *replay {
		// use timestamp to ensure uniqueness
		group = fmt.Sprintf("leaderboard-replay-%d", time.Now().UnixNano())
		slog.Info("replay mode: rebuilding view from offset 0", "group", group)
	}

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		// fresh group id
		"group.id": group,
		// since no offset to go off use earliest available
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		slog.Error("create consumer", "err", err)
		os.Exit(1)
	}
	defer c.Close()

	if err := c.SubscribeTopics([]string{topic}, nil); err != nil {
		slog.Error("subscribe", "err", err)
		os.Exit(1)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	tally := edit.NewTally() // the view: state = fold(Apply, events)
	var applied int
	lastPrint := time.Now()

	// loop forever
	for {
		// handle shutdown
		select {
		case <-sig:
			slog.Info("shutting down", "applied", applied)
			return
		default:
		}

		// Poll topic for 100ms
		ev := c.Poll(100)
		if ev == nil {
			maybePrint(tally, applied, &lastPrint)
			continue
		}

		switch m := ev.(type) {
		// handle message type
		case *kafka.Message:
			// parse
			e, perr := edit.ParseEdit(m.Value)
			if perr != nil {
				slog.Warn("unparseable message on log", "offset", m.TopicPartition.Offset, "err", perr)
				continue
			}
			tally.Apply(e) // 'fold' event to our tally
			applied++
			maybePrint(tally, applied, &lastPrint)
		// handle error
		case kafka.Error:
			slog.Error("kafka error", "err", m)
		}
	}
}

// called inside the loop so no mutex needed
func maybePrint(t *edit.Tally, applied int, last *time.Time) {
	if time.Since(*last) < 5*time.Second {
		return
	}
	*last = time.Now()

	fmt.Printf("\n=== top wikis · %d edits folded ===\n", applied)
	for i, wc := range t.TopWikis(10) {
		fmt.Printf("%2d. %-20s %d\n", i+1, wc.Wiki, wc.Count)
	}
}
