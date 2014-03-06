#compdef hk

# hk Autocomplete plugin for Oh-My-Zsh. Drop this plugin at
# ~/.oh-my-zsh/custom/plugins/hk/_hk to install it.
#
# Requires: The hk Heroku client (https://hk.heroku.com)
# Author: Blake Gentry (https://bgentry.io)

typeset -A opt_args

_hkrawcmds() {
  print -l ${(f)"$(hk help commands)"}
}

__hk_app_names() {
  # set a local curcontext to use for caching
  local curcontext=${curcontext%:*:*}:hk-__hk_app_names: state line cache_policy ret=1
  local cache_name=":completion:${curcontext}:"

  # See if a cache-policy is already set up, and set one if not
  zstyle -s $cache_name cache-policy cache_policy
  [[ -z "$cache_policy" ]] && zstyle $cache_name cache-policy _hk_app_names_caching_policy

  # If _app_names isn't populated or the cache is invalid, and we fail to
  # retrieve the cache:
  if ( ((${#_app_names} == 0)) || _cache_invalid $cache_name ) \
    && ! _retrieve_cache $cache_name; then
    # If we've gotten to this point, the app names aren't cached. Fetch them.
    _app_names=(${(f)"$(hk apps | cut -f 1 -d ' ')"})
    # Store _app_names in the cache
    _store_cache $cache_name _app_names
  fi

  compadd $* - $_app_names
}

_hk_app_names_caching_policy() {
  # Rebuild if cache is older than 1 hour.
  local -a oldp
  # This is a glob expansion for file modification time.
  # N sets NULL_GLOB, deleting the pattern from the arg list if it doesn't match.
  # m matches files with a given modification time, and h modifies the units to hours.
  # Finally, the +1 makes this match files modified at least 1 hour ago.
  oldp=( "$1"(Nmh+1) )
  # return the length of oldp (given by #)
  (( $#oldp ))
}

__hk_region_names() {
  # set a local curcontext to use for caching
  local curcontext=${curcontext%:*:*}:hk-__hk_region_names: state line cache_policy ret=1
  local cache_name=":completion:${curcontext}:"

  # See if a cache-policy is already set up, and set one if not
  zstyle -s $cache_name cache-policy cache_policy
  [[ -z "$cache_policy" ]] && zstyle $cache_name cache-policy _hk_region_names_caching_policy

  # If _region_names isn't populated or the cache is invalid, and we fail to
  # retrieve the cache:
  if ( ((${#_region_names} == 0)) || _cache_invalid $cache_name ) \
    && ! _retrieve_cache $cache_name; then
    # If we've gotten to this point, the region names aren't cached. Fetch them.
    _region_names=(${(f)"$(hk regions | cut -f 1 -d ' ')"})
    # Store _region_names in the cache
    _store_cache $cache_name _region_names
  fi

  compadd $* - $_region_names
}

_hk_region_names_caching_policy() {
  # Rebuild if cache is older than 2 weeks.
  local -a oldp
  # This is a glob expansion for file modification time.
  # N sets NULL_GLOB, deleting the pattern from the arg list if it doesn't match.
  # m matches files with a given modification time, and w modifies the units to weeks.
  # Finally, the +1 makes this match files modified at least 2 weeks ago.
  oldp=( "$1"(Nmw+2) )
  # return the length of oldp (given by #)
  (( $#oldp ))
}

_hkcmdnames() {
  print -l ${(f)"$(hk help commands | cut -f 2 -d ' ')"}
}

_hkcmdnames_needing_app() {
  print -l ${(f)"$(hk help commands | grep -F '[-a <app>]' | cut -f 2 -d ' ')"}
}

__args_for_command() {
  for i in _hkrawcmds; do
    if [[ "${${=i}[2]}" == $line[1] ]]; then
      local arglist; arglist=()
      #for j in (${^${i#*${line}[1] }}); do
      # print -l ${=${i#*${${line}[1]} }}
      for j in ${=${${i#*${${line}[1]} }%*\#*}}; do
        arglist+=($j)
      done
      _describe -t opts 'opts' arglist && ret=0

      # _arguments \
      #   '-a:appname'
      # ret=0
      return
    fi
  done
}

_hk-create() {
  local curcontext=$curcontext state line ret=1
  declare -A opt_args

  _arguments -C -s -S -A "-*" \
    '-r=[region]::heroku region name:__hk_region_names' \
    '*::app name:' \
   && ret=0

  return ret
}

_hk-env() {
  local curcontext=$curcontext state line ret=1
  declare -A opt_args

  # check if word is in array from _hkcmdnames_needing_app:
  # (${~${(j:|:)$(_hkcmdnames_needing_app)}})
  _arguments -C -s -S -A "-*" \
    '-a=[application name]:: :__hk_app_names' \
   && ret=0

  return ret
}

_hk-help() {
  local curcontext=$curcontext state line ret=1

  if (( CURRENT < 3 )); then
    compadd $(_hkcmdnames) && ret=0
  fi
  return ret
}

_hk() {
  integer ret=1
  local curcontext=$curcontext state line
  declare -A opt_args

  _arguments -C \
    '(-): :->command' \
    '*:: :->option-or-argument' \
  && return
    #'(-)*:: :->option-or-argument' \
  #   '2:help:->help' \
  #   '2:generators:->generator_lists' \
  #   '*:: :->args' \

  case "$state" in
    (command)
      # _describe -t commands 'hk command' hkcmds && ret=0
      # _describe -t commands 'help' hkcmdnames && ret=0
      #
      # _describe -t commands "hk command" cmdnames && ret=0
      #
      # compadd $(hk help commands | awk '{print $2}')

      #_describe_hkcmdnames
      compadd $(_hkcmdnames) && ret=0
    ;;
    (option-or-argument)
      local -a app_argument; app_argument='-a=[application name]:: :__hk_app_names'

      curcontext=${curcontext%:*:*}:hk-$words[1]:
      _call_function ret _hk-$words[1]
    ;;
  esac

  return ret
}

_hk
