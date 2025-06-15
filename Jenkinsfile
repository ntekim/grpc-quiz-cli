/* Requires the Docker Pipeline plugin */
pipeline {
    agent any

    tools {
        go 'go-latest'
    }

    environment {
        GOBIN = '/var/jenkins_home/go/bin'
        PATH = "${env.PATH}:${env.GOBIN}"
        GITHUB_TOKEN = credentials('Jenkins_PAT') // Add this token in Jenkins credentials
    }

    options {
        skipDefaultCheckout() // Manual checkout to use credentials
    }

    stages {
        stage('Checkout Code') {
            steps {
                script {
                    withCredentials([usernamePassword(
                        credentialsId: 'Jenkins_PAT', 
                        usernameVariable: 'GIT_USERNAME',
                        passwordVariable: 'GIT_PASSWORD'
                    )]) {
                        sh '''
                            git config --global credential.helper "store --file=.git-credentials"
                            git config --global user.name "ntekim"
                            git config --global user.email "jothamntekim@gmail.com"

                            if [ -d broker/.git ]; then
                                echo "Repo already exists, pulling latest changes"
                                cd broker
                                git checkout main
                                git pull origin main
                            else
                                echo "Cloning fresh copy of the repo"
                                git clone https://${GIT_USERNAME}:${GIT_PASSWORD}@github.com/nervix-ops/broker.git
                            fi
                        '''
                    }
                }
            }
        }

        stage('Check Go') {
            steps {
                sh '''
                    which go
                    go version
                '''
            }
        }

        stage('Build') {
            steps {
                sh 'cd broker && go mod download && go build -v ./...'
            }
        }

        stage('Test') {
            steps {
                sh 'cd broker && go test -v ./...'
            }
        }
        // stage('Lint (Optional)') {
        //     steps {
        //         sh '''
        //             cd broker
        //             if ! command -v golangci-lint > /dev/null; then
        //                 curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $GOBIN v1.56.2
        //             fi
        //             golangci-lint run
        //         '''
        //     }
        // }

        stage('Archive Build') {
            steps {
                archiveArtifacts artifacts: '**/*.go', allowEmptyArchive: true
            }
        }
        stage('deploy') {
            when {
              expression {
                env.BRANCH_NAME == 'dev' || env.BRANCH_NAME == 'main' || env.BRANCH_NAME == 'staging' 
              }
            }
            steps {
                sh 'go version'
                sh '''
                    echo "Multiline shell steps works too"
                    ls -lah
                    '''
            }
        }
        stage('cleanup') {
            steps {
                sh '''
                    echo "Multiline shell steps works too"
                    ls -lah
                    '''
                echo 'Pipeline ends here...'
            }
        }
    }
    post {
        success {
            echo '✅ Build completed successfully!'
        }
        failure {
            echo '❌ Build failed.'
        }
    }
}
