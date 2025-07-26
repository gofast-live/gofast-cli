# Script to run the stripe webhook listener
stripe listen --latest --forward-to localhost:4001/payments-webhook
