# Io - tui chatbot

I considered a few "platforms" for this project, discord bot, server hosted xterm bot. But in the interest of simplicity and project limitations I decided to do a tui.

##  features / roadmap

- [x] basic functionality
    - [x] tui UI
        - [x] input text
        - [x] chat log
        - [x] ascii image
        - [x] info panel
        - [x] status panel
    - [x] api requests (gemini only)
        - [x] standard request
        - [x] calls with conversation context

- [ ] submission ready
    - [ ] "runnable in 5 minutes" apikey considerations
    - [ ] test on other machines
    - [ ] "Your GitHub repo must have a README.md with an explanation of the project, what it does, how to run it, etc"
    - [ ] add disclaimer about partial use of AI to readme
    - [ ] social post

pref before turnin
- [ ] background colors
- [ ] dynamic size for ascii on manifestation
- [ ] create ai with colors
    - [ ] set it up as tool
    - [ ] tool should call the set ai command too, and print short introduction
    - [ ] tool call ai to make prompt
- [ ] emoji github.com/kyokomi/emoji
- [ ] openai
- [ ] changing terminal layout depending on aspect ratio

- [ ] features
    - [x] database
        - [x] tables/relations
            - [x] forgot ascii file path column
        - [x] CRUD
        - [x] integration into chat logic
    - [ ] UI
        - [ ] dynamic layout for wide terminal (currently basically assumes u in a tall terminal)
        - [ ] different color palettes (perhaps dynamically generated based on input image/ascii
        - [ ] newlines in textarea
        - [ ] chat
            - [ ] support emoji, both input & output, i'm thinking :thinking: for input
            - [ ] break up non streaming responses into blocks and send sentence by sentence or something like that, personalizes the "bot"
            - [ ] visual support for md, code snippets, etc.
        - [ ] status panel
            - [ ] beautify (animations, color, tons more spinners)
            - [ ] semi random messages (Thinking.. Processing..)
            - [ ] personalized state integration (depending on ai different messages)
            - [ ] emojis lol 🤔
        - [ ] info panel
            - [ ] coloring for different apis, like gemini blue, openai green
            - [ ] coloring for ai
            - [ ] list tokens used in this session/conversation, tokens used: int
            - [ ] ascii time
            - [ ] cava tool
    - [x] conversations
        - [x] /resume functionality, opens list of conversations in chatwindow, select one u want to set active
        - [x] /clear command to end conversation
        - [ ] /compact command to prevent conversation infinity growth, maybe force it
        - [ ] remove older conversations, currently it should infinitely expand
    - [ ] ais
        - [x] switch between ais functionality
        - [x] /show ais command
        - [x] /show apis command
        - [x] /show models (list all models in all apis) i'm thinking maybe combine these 2 for now
        - [x] add 2-3 on default database initialization
        - [x] /set ai command?
        - [x] /set prompt, api/model maybe
        - [ ] maybe a tool to select ai
            - [ ] maybe this should also call some request that sends introduction?
        - [x] create ai
            - [x] dynamically create ascii from filepath or imgur link, currently use terminal command, save in standard /ascii folder maybe
            - [ ] tool for ai to call to create ascii, with image link for ascii and prompt, defaults for api/model
    - [ ] memory
        - [ ] new table to store memories, foreign kei ai_id, CRUD
        - [ ] # command to remember like claude
        - [ ] compose prompt + memories on api requests
    - [ ] notifications / independence - some logic could maybe ping the user after randomized time or something if the program is left open
    - [ ] tools
        - [ ] image gen
        - [ ] web search
        - [ ] read webpage
    - [ ] API integration
        - [ ] OpenAI
        - [ ] XAI
        - [ ] Anthropic

    unrealistic...
    - [ ] mood, different moods could change prompt, and maybe use different emojis etc, the "ai" could then be dynamically changed under the hood based on switching mood levels, or maybe semi randomized
    - [ ] claude
        - [ ] tmux
        - [ ] hooks
    - [ ] fine tuned chatlogs
    - [ ] local inference
