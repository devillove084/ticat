# [Spec] Ticat command sequences

## Execute a sequence of command
A command sequence will execute commands one by one,
the latter one won't start untill the previous one finishes.
Commands in a sequence are seperated by ":".
```
$> ticat <command> : <command> : <command>

## Example:
$> ticat dummy : sleep 1s : echo hello

## Spaces(\s\t) are allowed but not necessary:
$> ticat dummy:sleep 1s:echo hello
```

## Display what will happen without execute a sequence
```
$> ticat <command> : <command> : <command> : desc

## Exmaples:
$> ticat dummy : desc
$> ticat dummy : sleep 1s : echo hello : desc
```

## Execute a sequence step by step
The env key "sys.step-by-step" enable or disable the step-by-step feature:
```
$> ticat {sys.step-by-step = true} <command> : <command> : <command>
$> ticat {sys.step-by-step = on} <command> : <command> : <command>
$> ticat {sys.step = on} <command> : <command> : <command>

## Enable it only for <command-2>, to ask for confirmation from user
$> ticat <command-1> : {sys.step = on} <command-2> : <command-3>
```

A set of builtin commands could changes this env key for better usage:
```
## Find these two commands:
$> ticat cmds.tree dbg.step
[step-by-step|step|s|S]
    - full-cmd:
        dbg.step-by-step
    - full-abbrs:
        dbg.step-by-step|step|s|S
    [on|yes|y|Y|1|+]
         'enable step by step'
        - full-cmd:
            dbg.step-by-step.on
        - full-abbrs:
            dbg.step-by-step|step|s|S.on|yes|y|Y|1|+
        - cmd-type:
            normal (quiet)
        - from:
            builtin
    [off|no|n|N|0|-]
         'disable step by step'
        - full-cmd:
            dbg.step-by-step.off
        - full-abbrs:
            dbg.step-by-step|step|s|S.off|no|n|N|0|-
        - cmd-type:
            normal (quiet)
        - from:
            builtin

## Use these commands:
$> ticat dbg.step.on : <command> : <command> : <command>

## Enable step-by-step in the middle
$> ticat <command> : <command> : dbg.step.on : <command>

## Enable and save, after this all executions will need confirming
$> ticat dbg.step.on : env.save
```

## The "desc" command branch

Overview
```
$> ticat cmds.tree.simple desc
[desc]
     'desc the flow about to execute'
    [simple]
         'desc the flow about to execute in lite style'
    [skeleton]
         'desc the flow about to execute, skeleton only'
    [dependencies]
         'list the depended os-commands of the flow'
    [env-ops-check]
         'desc the env-ops check result of the flow'
    [flow]
         'desc the flow execution'
        [simple]
             'desc the flow execution in lite style'
```

Exmaples of `desc`:
```
$> ticat <command> : <command> : <command> : desc

## Examples:
$> ticat dummy : desc
$> ticat dummy : sleep 1s : echo hello : desc
```

## Power/priority commands
Some commands have "power" flag, these type of command can changes the sequence.
Use "cmds.list <path>" or "cmds.tree <path>" can check a command's type.
```
## Example:
$> ticat cmds.tree dummy.power
[power|p|P]
     'power dummy cmd for testing'
    - full-cmd:
        dummy.power
    - full-abbrs:
        dummy|dmy|dm.power|p|P
    - cmd-type:
        power
```

The "desc" command have 3 flags:
* quiet: it would display in the executing sequence(the boxes)
* priority: it got to run first, then others could be executed.
* power: it can change the sequence about to execute.
```
## The command type of "desc"
$> ticat cmds.tree desc
[desc|d|D]
     'desc the flow about to execute'
    - cmd-type:
        power (quiet) (priority)

## The usage of "desc"
$> ticat <command> : <command> : <command> : desc
## The actual execute order
$> ticat desc : <command> : <command> : <command>
## The actual execution: "desc" remove all the commands after display the sequence's info
$> ticat desc [: <command> : <command> : <command>]
```

Other power commmands:
```
$> ticat cmd +
[more|+]
     'display rich info base on:
      * if in a sequence having
          * more than 1 other commands: show the sequence execution.
          * only 1 other command and
              * has no args and the other command is
                  * a flow: show the flow execution.
                  * not a flow: show the command or the branch info.
              * has args: find commands under the branch of the other command.
      * if not in a sequence and
          * has args: do global search.
          * has no args: show global help.'
    - cmd-type:
        power (quiet) (priority)
    - from:
        builtin
...
```

When have more than one priority commands in a sequence:
```
## User input
$> ticat <command-1> : <command-2> : <priority-command-a> : <priority-command-b>

## Actual execute order:
$> ticat <priority-command-a> : <priority-command-b> : <command-1> : <command-2>
```
