# Io - tui chatbot

I considered a few "platforms" for this project, discord bot, server hosted xterm bot. But in the interest of simplicity and project limitations I decided to do a tui.

## planned features (ai usage)
- [ ] basic functionality
    - [x] tui setup (never touched it) (10%)
        - [x] chat history
        - [x] image
        - [x] chat window
    - [x] api calls
        - [x] standard calls
        - [x] calls with conversation context (30%)
        - [x] streaming responses (gemini) (75%)
    - [ ] memory
        - [ ] session memory
        - [ ] permanent memory

- [ ] extended functionality
    - [ ] dynamic layout for whide terminal
    - [ ] notifications / independence
    - [ ] tools
        - [ ] utillity
            - [ ] image gen
            - [ ] web search
            - [ ] read webpage
        - [ ] config
            - [ ] model
            - [ ] nickname
            - [ ] status
                - [ ] prompt
                - [ ] show memory
    - [ ] claude
        - [ ] tmux
        - [ ] hooks
    - [ ] fine tuned chatlogs
