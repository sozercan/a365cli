package commands

import (
	"fmt"
)

// CompletionCmd generates shell completion scripts.
type CompletionCmd struct {
	Shell string `arg:"" enum:"bash,zsh,fish" help:"Shell type (bash, zsh, fish)"`
}

func (c *CompletionCmd) Run(ctx *Context) error {
	switch c.Shell {
	case "bash":
		fmt.Println(bashCompletion)
	case "zsh":
		fmt.Println(zshCompletion)
	case "fish":
		fmt.Println(fishCompletion)
	}
	return nil
}

const bashCompletion = `# a365 bash completion
_a365_complete() {
    local cur prev words cword
    _init_completion || return

    local commands="teams auth completion"
    local teams_commands="list get channels chats search"
    local channels_commands="list get messages post reply"
    local chats_commands="list get messages send send-self get-message"
    local auth_commands="login status logout"
    local global_flags="--json --plain --verbose --force --no-input --dry-run --client-id --tenant-id --version --help"

    case "${COMP_WORDS[1]}" in
        teams)
            case "${COMP_WORDS[2]}" in
                channels)
                    COMPREPLY=($(compgen -W "${channels_commands} ${global_flags}" -- "${cur}"))
                    ;;
                chats)
                    COMPREPLY=($(compgen -W "${chats_commands} ${global_flags}" -- "${cur}"))
                    ;;
                *)
                    COMPREPLY=($(compgen -W "${teams_commands} ${global_flags}" -- "${cur}"))
                    ;;
            esac
            ;;
        auth)
            COMPREPLY=($(compgen -W "${auth_commands} ${global_flags}" -- "${cur}"))
            ;;
        completion)
            COMPREPLY=($(compgen -W "bash zsh fish" -- "${cur}"))
            ;;
        *)
            COMPREPLY=($(compgen -W "${commands} ${global_flags}" -- "${cur}"))
            ;;
    esac
}
complete -F _a365_complete a365`

const zshCompletion = `#compdef a365

_a365() {
    local -a commands
    commands=(
        'teams:Microsoft Teams'
        'auth:Authentication'
        'completion:Generate shell completion script'
    )

    local -a teams_commands
    teams_commands=(
        'list:List joined teams'
        'get:Get a team by ID'
        'channels:Team channels'
        'chats:Team chats'
        'search:Search Teams messages'
    )

    local -a channels_commands
    channels_commands=(
        'list:List channels in a team'
        'get:Get a channel by ID'
        'messages:List messages in a channel'
        'post:Post a message to a channel'
        'reply:Reply to a channel message'
    )

    local -a chats_commands
    chats_commands=(
        'list:List recent chats'
        'get:Get a chat by ID'
        'messages:List messages in a chat'
        'send:Send a message to a chat'
        'send-self:Send a message to yourself'
        'get-message:Get a specific message from a chat'
    )

    local -a auth_commands
    auth_commands=(
        'login:Log in via browser'
        'status:Show authentication status'
        'logout:Log out and clear credentials'
    )

    _arguments -C \
        '--json[Output as JSON]' \
        '--plain[Output as plain TSV]' \
        '(-v --verbose)'{-v,--verbose}'[Show MCP request/response]' \
        '--force[Skip confirmation prompts]' \
        '--no-input[Never prompt]' \
        '--dry-run[Preview write operations]' \
        '--client-id=[Entra app client ID]:client id' \
        '--tenant-id=[Entra tenant ID]:tenant id' \
        '(-V --version)'{-V,--version}'[Show version]' \
        '1: :->cmd' \
        '*:: :->args'

    case $state in
        cmd)
            _describe 'command' commands
            ;;
        args)
            case $words[1] in
                teams)
                    case $words[2] in
                        channels) _describe 'command' channels_commands ;;
                        chats) _describe 'command' chats_commands ;;
                        *) _describe 'command' teams_commands ;;
                    esac
                    ;;
                auth) _describe 'command' auth_commands ;;
                completion) compadd bash zsh fish ;;
            esac
            ;;
    esac
}

_a365`

const fishCompletion = `# a365 fish completion
complete -c a365 -n '__fish_use_subcommand' -a teams -d 'Microsoft Teams'
complete -c a365 -n '__fish_use_subcommand' -a auth -d 'Authentication'
complete -c a365 -n '__fish_use_subcommand' -a completion -d 'Generate shell completion'

# teams subcommands
complete -c a365 -n '__fish_seen_subcommand_from teams; and not __fish_seen_subcommand_from list get channels chats search' -a list -d 'List joined teams'
complete -c a365 -n '__fish_seen_subcommand_from teams; and not __fish_seen_subcommand_from list get channels chats search' -a get -d 'Get a team by ID'
complete -c a365 -n '__fish_seen_subcommand_from teams; and not __fish_seen_subcommand_from list get channels chats search' -a channels -d 'Team channels'
complete -c a365 -n '__fish_seen_subcommand_from teams; and not __fish_seen_subcommand_from list get channels chats search' -a chats -d 'Team chats'
complete -c a365 -n '__fish_seen_subcommand_from teams; and not __fish_seen_subcommand_from list get channels chats search' -a search -d 'Search messages'

# channels subcommands
complete -c a365 -n '__fish_seen_subcommand_from channels; and not __fish_seen_subcommand_from list get messages post reply' -a list -d 'List channels'
complete -c a365 -n '__fish_seen_subcommand_from channels; and not __fish_seen_subcommand_from list get messages post reply' -a get -d 'Get a channel'
complete -c a365 -n '__fish_seen_subcommand_from channels; and not __fish_seen_subcommand_from list get messages post reply' -a messages -d 'List messages'
complete -c a365 -n '__fish_seen_subcommand_from channels; and not __fish_seen_subcommand_from list get messages post reply' -a post -d 'Post a message'
complete -c a365 -n '__fish_seen_subcommand_from channels; and not __fish_seen_subcommand_from list get messages post reply' -a reply -d 'Reply to a message'

# chats subcommands
complete -c a365 -n '__fish_seen_subcommand_from chats; and not __fish_seen_subcommand_from list get messages send send-self get-message' -a list -d 'List chats'
complete -c a365 -n '__fish_seen_subcommand_from chats; and not __fish_seen_subcommand_from list get messages send send-self get-message' -a get -d 'Get a chat'
complete -c a365 -n '__fish_seen_subcommand_from chats; and not __fish_seen_subcommand_from list get messages send send-self get-message' -a messages -d 'List messages'
complete -c a365 -n '__fish_seen_subcommand_from chats; and not __fish_seen_subcommand_from list get messages send send-self get-message' -a send -d 'Send a message'
complete -c a365 -n '__fish_seen_subcommand_from chats; and not __fish_seen_subcommand_from list get messages send send-self get-message' -a send-self -d 'Send to yourself'
complete -c a365 -n '__fish_seen_subcommand_from chats; and not __fish_seen_subcommand_from list get messages send send-self get-message' -a get-message -d 'Get a message'

# auth subcommands
complete -c a365 -n '__fish_seen_subcommand_from auth' -a login -d 'Log in'
complete -c a365 -n '__fish_seen_subcommand_from auth' -a status -d 'Show status'
complete -c a365 -n '__fish_seen_subcommand_from auth' -a logout -d 'Log out'

# completion subcommands
complete -c a365 -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish'

# global flags
complete -c a365 -l json -d 'Output as JSON'
complete -c a365 -l plain -d 'Output as plain TSV'
complete -c a365 -s v -l verbose -d 'Show MCP request/response'
complete -c a365 -l force -d 'Skip confirmation prompts'
complete -c a365 -l no-input -d 'Never prompt'
complete -c a365 -l dry-run -d 'Preview write operations'
complete -c a365 -l client-id -d 'Entra app client ID' -x
complete -c a365 -l tenant-id -d 'Entra tenant ID' -x`
