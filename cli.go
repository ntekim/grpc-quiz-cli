package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	pb "github.com/ntekim/grpc-cli-quiz/proto"
	"github.com/ntekim/grpc-cli-quiz/server"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	client               pb.CLIQuizServiceClient
	currentQuestionIndex = 0
	questions            []*pb.Question
)

var (
	answers map[int32]*pb.Answer
	results map[int32]*pb.ResultResponsePayload
	port    string
)

func getNextQuestion(ctx context.Context) {
	if len(questions) == 0 {
		fmt.Println("No questions to display.")
		return
	}
	for currentQuestionIndex < len(questions) {
		select {
		case <-ctx.Done():
			log.Println("Exiting question loop...")
			return
		default:
			question := questions[currentQuestionIndex]
			shuffleOptions(question)
			fmt.Printf("\nQ%d: %s\n", question.GetId(), question.GetQuestionDesc())
			for i, option := range question.GetOptions() {
				fmt.Printf("%d) %s\n", i+1, option)
			}

			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter your answer (number): ")
			input, err := reader.ReadString('\n')
			fmt.Println()
			if err != nil {
				fmt.Printf("Failed to read input: %v\n", err)
				return
			}
			input = strings.TrimSpace(input)
			answerIndex, err := strconv.Atoi(input)
			if err != nil || answerIndex < 1 || answerIndex > len(question.GetOptions()) {
				fmt.Println("Invalid input. Please enter a valid option number.")
				continue
			}

			if answers[question.GetId()] == nil {
				answers[question.GetId()] = &pb.Answer{}
			}

			answers[question.GetId()].Answer = question.GetOptions()[answerIndex-1]

			currentQuestionIndex++
		}
	}

	fmt.Println("You have answered all questions. Submitting answers...")
	submitAnswers(ctx)
}

func submitAnswers(ctx context.Context) {
	select {
	case <-ctx.Done():
		log.Println("Exiting question loop...")
		return
	default:
		fmt.Println("\nSubmitted your answers...")
		userID := generateUUID()

		answerList := []*pb.Answer{}
		for qID, ans := range answers {
			answerList = append(answerList, &pb.Answer{QuestionId: qID, Answer: ans.GetAnswer()})
		}
		req := &pb.AnswersRequestPayload{Answers: answerList, UserId: userID}

		res, err := client.SubmitAnswers(ctx, req)
		if err != nil {
			fmt.Printf("Failed to submit answers: %v\n", err)
			return
		}
		resultMsg := compareResults(res)

		fmt.Printf("You got %d correct answers!\n", res.GetCorrectAnswerCount())
		fmt.Println(resultMsg)

		retakeQuiz(ctx)

	}
}

func retakeQuiz(ctx context.Context) {
	fmt.Println(results)
	fmt.Println()
	fmt.Println("Retake Quiz?")
	fmt.Println("1) Yes")
	fmt.Println("2) No")

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your answer (number): ")
	input, err := reader.ReadString('\n')
	fmt.Println()
	if err != nil {
		fmt.Printf("Failed to read input: %v\n", err)
		return
	}
	input = strings.TrimSpace(input)
	answerIndex, err := strconv.Atoi(input)
	if err != nil || answerIndex != 2 && answerIndex != 1 {
		fmt.Println("Invalid input. Please enter a valid option number.")
		retakeQuiz(ctx)
	}
	switch answerIndex {
	case 1:
		fmt.Println("Retaking quiz...")
		currentQuestionIndex = 0
		answers = make(map[int32]*pb.Answer)
		getNextQuestion(ctx)
	case 2:
		fmt.Println("Exiting quiz...")
		os.Exit(0)
	}
}

func compareResults(res *pb.ResultResponsePayload) string {
	var resultMsg string
	resultLength := len(results)
	if resultLength <= 0 {
		resultMsg = "You're the first to attempt this quiz."
		results[int32(0)] = res
	} else {
		count := 0
		for _, result := range results {
			if res.GetResultPercentage() > result.GetResultPercentage() {
				count++
			}
		}
		results[int32(resultLength)+int32(1)] = res
		resultLength = len(results)
		percentageMultiplier := 100
		userPercentagePerformance := float32(count) / float32(resultLength) * float32(percentageMultiplier)
		resultMsg = fmt.Sprintf("You were better than %.0f%% of all quizzers.", float32(userPercentagePerformance))
	}

	return resultMsg
}

func generateUUID() string {
	userID := uuid.New()

	return userID.String()
}

func shuffleQuestions(questions []*pb.Question) []*pb.Question {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Shuffle(len(questions), func(i, j int) {
		questions[i], questions[j] = questions[j], questions[i]
	})
	return questions
}

func shuffleOptions(question *pb.Question) {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Shuffle(len(question.GetOptions()), func(i, j int) {
		question.GetOptions()[i], question.GetOptions()[j] = question.GetOptions()[j], question.GetOptions()[i]
	})
}

func InitClient(port string) error {
	conn, err := grpc.NewClient("127.0.0.1:"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect: %v", err)
		return err
	}
	client = pb.NewCLIQuizServiceClient(conn)
	answers = make(map[int32]*pb.Answer)
	results = make(map[int32]*pb.ResultResponsePayload)
	return nil
}

var quizCmd = &cobra.Command{
	Use:   "start-quiz",
	Short: "Start the quiz and answer questions one by one",
	Run: func(cmd *cobra.Command, _ []string) {
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			server.NewServer(cmd.Context(), &wg, port)
		}()

		time.Sleep(2 * time.Second)

		go func() {
			<-cmd.Context().Done()
			log.Println("Quiz interrupted, shutting down gracefully...")
		}()

		if err := InitClient(port); err != nil {
			log.Printf("Failed to initialize client: %v\n", err)
			return
		}

		res, err := client.GetQuestions(cmd.Context(), &pb.NoRequestParam{})
		if err != nil {
			log.Printf("Failed to get questions: %v\n", err)
			return
		}

		if res.GetQuestions() == nil || len(res.GetQuestions()) == 0 {
			fmt.Println("No questions received from the server.")
			return
		}
		shuffledQuestions := shuffleQuestions(res.GetQuestions())
		questions = shuffledQuestions

		getNextQuestion(cmd.Context())
		wg.Wait()
	},
}

func StartCLI(ctx context.Context) error {
	fmt.Println("Starting Quiz...")
	rootCmd := &cobra.Command{Use: "quiz-cli"}
	quizCmd.PersistentFlags().StringVarP(&port, "port", "p", "50051", "Server port")
	rootCmd.AddCommand(quizCmd)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
