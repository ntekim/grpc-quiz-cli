package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	pb "github.com/ntekim/grpc-cli-quiz/proto"
	"github.com/spf13/cobra"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var client pb.CLIQuizServiceClient
var currentQuestionIndex = 0
var questions []*pb.Question

var answers map[int32]*pb.Answer

func InitClient() {
	conn, err := grpc.NewClient("127.0.0.1:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect: %v", err)
		os.Exit(1)
	}
	client = pb.NewCLIQuizServiceClient(conn)
	answers = make(map[int32]*pb.Answer)
}

func getNextQuestion() {
	if len(questions) == 0 {
		fmt.Println("No questions to display.")
		return
	}
	for currentQuestionIndex < len(questions) {
		question := questions[currentQuestionIndex]
		fmt.Printf("\nQ%d: %s\n", question.Id, question.GetQuestionDesc())
		for i, option := range question.GetOptions() {
			fmt.Printf("%d) %s\n", i+1, option)
		}

		// Get user input for this question
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your answer (number): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Failed to read input: %v\n", err)
			return
		}
		input = strings.TrimSpace(input)

		// Convert input to integer and validate
		answerIndex, err := strconv.Atoi(input)
		if err != nil || answerIndex < 1 || answerIndex > len(question.GetOptions()) {
			fmt.Println("Invalid input. Please enter a valid option number.")
			continue
		}

		// Check if the entry for question.Id exists, and initialize it if not
		if answers[question.GetId()] == nil {
			answers[question.GetId()] = &pb.Answer{} // Initialize with an empty Answer struct
		}

		answers[question.GetId()].Answer = question.Options[answerIndex-1]

		// Move to the next question
		currentQuestionIndex++
	}

	fmt.Println("You have answered all questions. Submitting answers...")
	submitAnswers() // Submit the answers after all questions are answered
}

func submitAnswers() {
	fmt.Println("\nSubmitting your answers...")

	answerList := []*pb.Answer{}
	for qID, ans := range answers {
		answerList = append(answerList, &pb.Answer{QuestionId: qID, Answer: ans.GetAnswer()})
	}
	req := &pb.AnswersRequestPayload{Answers: answerList}

	res, err := client.SubmitAnswers(context.Background(), req)
	if err != nil {
		fmt.Printf("Failed to submit answers: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("You got %d correct answers!\n", res.GetCorrectAnswerCount())
}

var quizCmd = &cobra.Command{
	Use:   "start-quiz",
	Short: "Start the quiz and answer questions one by one",
	Run: func(cmd *cobra.Command, args []string) {
		InitClient()

		res, err := client.GetQuestions(context.Background(), &pb.NoRequestParam{})
		if err != nil {
			fmt.Printf("Failed to get questions: %v\n", err)
			os.Exit(1)
		}

		if res.GetQuestions() == nil || len(res.GetQuestions()) == 0 {
			fmt.Println("No questions received from the server.")
			return
		}

		questions = res.GetQuestions()
		fmt.Printf("Number of questions fetched: %d\n", len(questions))

		getNextQuestion()
	},
}

func StartCLI() {
	var rootCmd = &cobra.Command{Use: "quiz-cli"}
	rootCmd.AddCommand(quizCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
