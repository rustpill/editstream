package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/rustpill/editstream/internal/edit"
	"github.com/rustpill/editstream/internal/stream"
)

const topic = "edits"

func main() {
	// cancelable context derived from context.Background()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// create producer
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		// first contact address
		"bootstrap.servers": "localhost:9092",
	})

	if err != nil {
		slog.Error("create producer", "err", err)
		os.Exit(1)
	}

	defer p.Close()

	go func() {
		// internal producer event handler
		for e := range p.Events() {
			if m, ok := e.(*kafka.Message); ok && m.TopicPartition.Error != nil {
				slog.Error("delivery failed", "err", m.TopicPartition.Error)
			}
		}
	}()

	var produced, skipped int
	// create Client object
	client := &stream.Client{
		URL:       stream.DefaultURL,
		UserAgent: "editstream/1.0 (github.com/rustpill/editstream)",
	}

	err = client.Run(ctx, func(data []byte) {
		// for every edit

		// parse data into our Edit struct
		e, perr := edit.ParseEdit(data)
		if perr != nil {
			skipped++
			slog.Debug("skipping unparseable event", "err", perr)
			return
		}

		t := topic
		// send to topic
		if err := p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &t, Partition: kafka.PartitionAny},
			Key:            []byte(e.Wiki),
			Value:          data,
		}, nil); err != nil {
			slog.Error("produce", "err", err)
			return
		}
		produced++
		// feedback every 100 events
		if produced%100 == 0 {
			slog.Info("progress", "produced", produced, "skipped", skipped)
		}
	})

	slog.Info("shutting down", "reason", err, "produced", produced, "skipped", skipped)
	p.Flush(5000) // wait 5 seconds before shut down

}
