#!/usr/bin/env python3
"""
Basic usage example for Modelplex with OpenAI client
"""

import openai
import os

def main():
    # Configure the OpenAI client to use Modelplex
    client = openai.OpenAI(
        base_url="http://unix:/path/to/modelplex.socket",
        api_key="unused"  # API key is not needed, handled by Modelplex
    )
    
    print("🔗 Connected to Modelplex")
    
    # List available models
    try:
        models = client.models.list()
        print(f"📋 Available models: {[model.id for model in models.data]}")
        
        # Use GPT-4 through Modelplex
        response = client.chat.completions.create(
            model="gpt-4",
            messages=[
                {"role": "system", "content": "You are a helpful assistant running in complete network isolation through Modelplex."},
                {"role": "user", "content": "Hello! Can you explain what Modelplex does?"}
            ],
            max_tokens=200
        )
        
        print("🤖 GPT-4 Response:")
        print(response.choices[0].message.content)
        
        # Try Claude through the same interface
        claude_response = client.chat.completions.create(
            model="claude-3-sonnet", 
            messages=[
                {"role": "user", "content": "What are the advantages of running AI in isolation?"}
            ],
            max_tokens=150
        )
        
        print("\n🧠 Claude Response:")
        print(claude_response.choices[0].message.content)
        
    except Exception as e:
        print(f"❌ Error: {e}")
        print("Make sure Modelplex is running and the socket path is correct")

if __name__ == "__main__":
    main()