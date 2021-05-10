package cmd

import (
	"log"
	"os"
	"os/signal"
	"regexp"

	"github.com/gosuri/uiprogress"
	"github.com/nazar256/amqp-message-remover/remover"
	"github.com/spf13/cobra"
)

const (
	minArgLength = 2
	maxArgLength = 3
)

var (
	continuous   bool
	nack         bool
	prefetch     int
	matchHeaders bool
)

var rootCmd = &cobra.Command{
	Use: `amqp-message-remover [queue_name] [regex] [DSN] [options]
regex - regular expression without delimiters (example: .*some-bad-value.*).
DSN - amqp url specification, default is amqp://guest:guest@127.0.0.1:5672
`,
	Args:  cobra.RangeArgs(minArgLength, maxArgLength),
	Short: "Removes messages from AMQP queue selectively by regexp",
	Long: `Removes messages from AMQP queue selectively. Matches message body by default against regular expression.
Only unacked messages can be processed. 
If you have to process all your queue - prefetch count must be greater or equal than actual queue size.
Alternatively you can run this command in parallel with your normal consumer. In this case this command will remove
messages not consumed by another consumer.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		dsn := "amqp://guest:guest@127.0.0.1:5672"
		if len(args) >= maxArgLength {
			dsn = args[2]
		}

		interruptCh := make(chan os.Signal, 1)
		signal.Notify(interruptCh, os.Interrupt)
		defer func() {
			signal.Stop(interruptCh)
		}()

		bar := makeBar()

		config := remover.Config{
			Dsn:           dsn,
			QueueName:     args[0],
			PrefetchCount: uint16(prefetch),
			Regexp:        regexp.MustCompile(args[1]),
			MatchType:     remover.MatchBody,
			Continuous:    continuous,
		}
		if matchHeaders {
			config.MatchType = remover.MatchHeaders
		}
		statusCh := remover.RemoveMessages(config)

		for {
			select {
			case status := <-statusCh:
				progressValue := status.Processed
				if continuous {
					progressValue = status.Removed
				}
				err = bar.Set(progressValue)

				if err != nil {
					return err
				}
				if status.Finished {
					return
				}
				if !continuous && status.Processed >= prefetch {
					return
				}
			case <-interruptCh:
				return
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
	log.Println("done")
}

func init() {
	rootCmd.Flags().IntVar(
		&prefetch,
		"prefetch",
		1,
		"prefetch count, how many last messages to scan (or minimum to delete when 'continuous');",
	)

	_ = rootCmd.MarkFlagRequired("prefetch")

	rootCmd.Flags().BoolVar(
		&continuous,
		"continuous",
		false,
		"use this flag when running along with regular consumer, this is the same as if you rerun command continuously",
	)

	rootCmd.Flags().BoolVar(&nack, "nack", false, "use to unacknowledge messages, i.e. to send them to DLX")

	rootCmd.Flags().BoolVar(
		&matchHeaders,
		"headers",
		false,
		"match headers instead of body, they will be serialized to json before matching",
	)
}

func makeBar() *uiprogress.Bar {
	uiprogress.Start()
	bar := uiprogress.AddBar(prefetch)
	bar.AppendCompleted()

	return bar
}
