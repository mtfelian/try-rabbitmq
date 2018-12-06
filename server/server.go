package main

import (
	"encoding/json"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"os"
)

func configure() error {
	viper.SetConfigFile("config.yaml")
	return viper.ReadInConfig()
}

func initialize(channel *amqp.Channel, exchange, qIn, qOut string) error {
	if err := channel.ExchangeDeclare(exchange, "topic", false, false, false, false, nil); err != nil {
		return err
	}

	if _, err := channel.QueueDeclare(qIn, false, false, false, false, nil); err != nil {
		return err
	}

	if _, err := channel.QueueDeclare(qOut, false, false, false, false, nil); err != nil {
		return err
	}

	if err := channel.QueueBind(qIn, qIn, exchange, false, nil); err != nil {
		return err
	}
	if err := channel.QueueBind(qOut, qOut, exchange, false, nil); err != nil {
		return err
	}

	return nil
}

// evaluate calculates the formula for the given val
func evaluate(formula string) (float64, error) {
	expr, err := govaluate.NewEvaluableExpression(formula)
	if err != nil {
		return 0, fmt.Errorf("failed to compile formula %s: %v", formula, err)
	}

	data := map[string]interface{}{}
	r, err := expr.Evaluate(data)
	if err != nil {
		return 0, fmt.Errorf("failed to evaluate formula %s (data: %#v): %v", formula, data, err)
	}

	rFloat64, ok := r.(float64)
	if !ok {
		return 0, fmt.Errorf("result is not float64 for (data: %#v), formula %s", data, formula)
	}
	return rFloat64, nil
}

type result struct {
	ID   string  `json:"id"`
	Data float64 `json:"data"`
	Err  error   `json:"err"`
	OK   bool    `json:"ok"`
}

func startPerformTasks(channel *amqp.Channel, exchange, qTask, qResults string) error {
	resultsC, err := channel.Consume(qTask, "", true, true, false, false, nil)
	if err != nil {
		return err
	}

	logrus.Infoln("Starting main loop...")
	for {
		msg, ok := <-resultsC
		if !ok {
			logrus.Errorln("Channel resultC is closed")
			os.Exit(1)
		}

		var resultStruct result
		r, err := evaluate(string(msg.Body))
		resultStruct = result{OK: true, ID: msg.MessageId, Data: r}
		if err != nil {
			resultStruct = result{Err: err, OK: false, ID: msg.MessageId}
		}

		resultJSON, err := json.Marshal(resultStruct)
		if err != nil {
			logrus.Errorf("failed to json.Marshal: %v", err)
			continue
		}

		logrus.Infof("Publishing result: %s", string(resultJSON))
		if err := channel.Publish(exchange, qResults, false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        resultJSON,
		}); err != nil {
			logrus.Errorln("failed to Publish: %v", err)
			continue
		}

		logrus.Infof("Result: %s", string(msg.Body))
	}
	return nil
}

func main() {
	if err := configure(); err != nil {
		logrus.Fatal(err)
	}

	rmqDialString := viper.GetString("rabbitmq")
	rmqExchange := viper.GetString("rmq_exchange")
	rmqKeyTasks := viper.GetString("rmq_key_tasks")
	rmqKeyResults := viper.GetString("rmq_key_results")

	conn, err := amqp.Dial(rmqDialString)
	if err != nil {
		logrus.Fatal(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		logrus.Fatal(err)
	}
	defer func() {
		if err := ch.Close(); err != nil {
			logrus.Errorf("Channel close error: %v", err)
		}
	}()

	if err := initialize(ch, rmqExchange, rmqKeyTasks, rmqKeyResults); err != nil {
		logrus.Fatal(err)
	}

	if err := startPerformTasks(ch, rmqExchange, rmqKeyTasks, rmqKeyResults); err != nil {
		logrus.Fatal(err)
	}
}
