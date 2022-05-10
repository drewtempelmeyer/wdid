# wdid

What did I do?/What do I do? Keeps track of your tasks in the terminal. Very
much an early, functioning prototype.


## Installation

### Install using Go
```
go install github.com/drewtempelmeyer/wdid
```

### Helpful Aliases (optional)

Typing `wdid` before every command can be a pain. You can add the following
aliases in your `.bashrc` or `.zshrc`:
```bash
alias did="wdid did"
alias todo="wdid do"
```

## Usage

### Logging a completed task (`wdid did`)

When you've completed a task that's not currently tracked using wdid, you can
create the task using `wdid did`:
```bash
wdid did Added usage instructions to the README
```

The task will be added and immediately marked as completed.

### Adding a task to be done (`wdid do`)

This command creates an entry to be done and will not be marked as completed.
This is useful for planning out your day or working on long-lived tasks. The
standup report will show you worked on this the day the entry was created.

```
wdid do Complete the README instructions
```

### Completing tasks (`wdid complete`)

Marks the task(s) as completed. To specify the task(s), pass the id (refer to
`wdid standup` to retrieve the id numbers) or multiple id numbers, separated by
a space.

```
wdid complete 20 32
```

### Deleting tasks (`wdid delete`)

Deletes the task(s) specified by the id number(s) passed to the command. This
cannot be undone.

```
wdid delete 20 32
```

### Generating a standup report (`wdid standup`)

Outputs a Markdown report of yesterday's and today's tasks.

```
wdid standup
```

### Backing up your tasks

`wdid` stores everything locally in a SQLite database. To create a backup, copy
the `.wdid-db` file in your home directory.
