#!/usr/bin/env python3

import aws_cdk as cdk

from mqcdk.mqcdk_stack import MqcdkStack

app = cdk.App()
MqcdkStack(app, "MqcdkStack")

app.synth()
