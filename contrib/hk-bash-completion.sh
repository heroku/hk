#!/bin/bash

_hk_commands()
{
    hk help commands|cut -f 2 -d ' '
}

_hk()
{
    cur=${COMP_WORDS[COMP_CWORD]}
    prev=${COMP_WORDS[COMP_CWORD-1]}
    if [ $COMP_CWORD -eq 1 ]; then
        COMPREPLY=( $( compgen -W "$(_hk_commands)" $cur ) )
    elif [ $COMP_CWORD -eq 2 ]; then
        case "$prev" in
        help)
            COMPREPLY=( $( compgen -W "$(_hk_commands)" $cur ) )
            ;;
        esac
    fi
}

complete -F _hk -o default hk
