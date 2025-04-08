import aws_cdk as core
import aws_cdk.assertions as assertions

from mqcdk.mqcdk_stack import MqcdkStack

# example tests. To run these tests, uncomment this file along with the example
# resource in mqcdk/mqcdk_stack.py
def test_sqs_queue_created():
    app = core.App()
    stack = MqcdkStack(app, "mqcdk")
    template = assertions.Template.from_stack(stack)

#     template.has_resource_properties("AWS::SQS::Queue", {
#         "VisibilityTimeout": 300
#     })
