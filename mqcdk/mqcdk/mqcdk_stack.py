from aws_cdk import (
    CfnOutput,
    Duration,
    Stack,
    aws_sqs,
)
from constructs import Construct


class MqcdkStack(Stack):
    def __init__(self, scope: Construct, construct_id: str, **kwargs) -> None:
        super().__init__(scope, construct_id, **kwargs)
        sqs_queue_name = "main"

        self.dead_letter_queue = aws_sqs.Queue(
            self, "MyDeadLetterQueue", queue_name=f"{sqs_queue_name}-dlq"
        )

        self.main_queue = aws_sqs.Queue(
            self,
            "MyMainQueue",
            queue_name=sqs_queue_name,
            delivery_delay=Duration.seconds(2),
            max_message_size_bytes=2048,
            retention_period=Duration.days(1),
            receive_message_wait_time=Duration.seconds(10),
            visibility_timeout=Duration.minutes(5),
            dead_letter_queue=aws_sqs.DeadLetterQueue(
                queue=self.dead_letter_queue, max_receive_count=2
            ),
        )

        CfnOutput(self, "MainQueueUrl", value=self.main_queue.queue_url)
        CfnOutput(self, "DeadLetterQueueUrl", value=self.dead_letter_queue.queue_url)
        CfnOutput(self, "MainQueueArn", value=self.main_queue.queue_arn)
        CfnOutput(self, "DeadLetterQueueArn", value=self.dead_letter_queue.queue_arn)
