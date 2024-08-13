pipeline {
    agent any

    environment {
        // AWS environment variables
        AWS_DEFAULT_REGION = 'us-west-2'
        AWS_ACCESS_KEY_ID = credentials('aws-access-key-id') // Stored in Jenkins Credentials
        AWS_SECRET_ACCESS_KEY = credentials('aws-secret-access-key') // Stored in Jenkins Credentials
        S3_BUCKET = 'your-s3-bucket-name'
        IMAGE_REPO_NAME = 'your-ecr-repo-name'
        IMAGE_TAG = "latest"
        ECR_REGISTRY = "your-aws-account-id.dkr.ecr.${AWS_DEFAULT_REGION}.amazonaws.com"
    }

    stages {
        stage('Checkout Code') {
            steps {
                checkout scm
            }
        }

        stage('Build') {
            steps {
                sh 'go mod download'
                sh 'go build -o go-dynamodb-crud'
            }
        }

        stage('Test') {
            steps {
                script {
                    sh 'go test -v ./...'
                }
            }
        }

        stage('Docker Build & Push') {
            steps {
                script {
                    sh "aws ecr get-login-password --region ${AWS_DEFAULT_REGION} | docker login --username AWS --password-stdin ${ECR_REGISTRY}"
                    sh "docker build -
