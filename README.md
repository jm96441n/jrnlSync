# jrnlSync
Simple tool to backup notes from [jrnl](https://jrnl.sh/en/stable/) to another provider. Currently only works on unix systems with cron as it
leverages cron to handle the automatic backups, though the sync command should work on windows as well (not tested
though).

## Feature/Provider List:
- [ ] Custom cron setting for sync timing
- [ ] [Notion](https://www.notion.so)
    - [X] Daily sync of previous day notes
    - [ ] Initial sync of all notes
    - [ ] Sync from multiple machines (don't overwrite existing page)
- [ ] Obisidan
- [ ] Roam
- [ ] Timeline

## Installation

* Homebrew
* DNF/YUM
* APT/DEB
* go install

### Homebrew

```
brew tap jm96441n/jrnlSync
brew install jrnlSync
```

### DNF/YUM

1. Add the repository by creating the following file `/etc/yum.repos.d/jrnlSync.repo`, with:
```
 [jrnlSync]
name=jrnlSync Repository
baseurl=https://yum.fury.io/jm96441n
enabled=1
gpgcheck=0
```

2. Install:
```
sudo dnf install jrnlSync
```

### APT/DEB

1. Add the repository by creating the following file `/etc/apt/sources.list.d/jrnlSync.list`, with:
```
deb [trusted=yes] https://apt.fury.io/jm96441n/ /
```

2. Instal:
```
sudo apt-get install jrnlSync
```

### go install

If you have the go toolchain and prefer to install go binaries that way you can install with:
```
go install https://github.com/jm96441n/jrnlSync@latest
```


## Usage

The following commands are available
`setup`
`notion`

NOTE: Right now only the body of the notes are synced, soon we'll be syncing both the title and body.

### `setup`

This command should be used the first time you download the package, as it's used to set the cron job that will perform
nightly syncs of your jrnl entries. To set this up you need a [Notion Integration Key](https://developers.notion.com/docs/getting-started#step-1-create-an-integration) and
you'll need the id of your notion database that these entries will be backed up to, instructions from notion [here](https://developers.notion.com/docs/working-with-databases#adding-pages-to-a-database)
or from the notion documentation:
> Here's a quick procedure to find the database ID for a specific database in Notion:

    Open the database as a full page in Notion. Use the Share menu to Copy link. Now paste the link in your text editor so you can take a closer look. The URL uses the following format:

    `https://www.notion.so/{workspace_name}/{database_id}?v={view_id}`

Once you have both of those keys, just run `jrnlSync setup` and it will prompt you for the information needed. (Note we
don't store these keys anywhere, they're written to the crontab as arguments to the `notion` command).

And that's it! Now every night right after midnight the `notion` command will be run which will sync your notes from the
day prior to a page in your notion database.

### `notion`

This command is to sync notes to notion. As of now this command only handles syncing notes from the day prior to a page
in a notion database. You can call this command with:

```
jrnlSync notion -d [DATABASE_ID] -k [NOTION_INTEGRATION_KEY]
```

where `[DATABASE_ID]` is replaced with the database to write the pages to and `[NOTION_INTEGRATION_KEY]` is replaced
with your notion integration key.


## Contributing

Feel free to fork the repo and the open a pull request! I'm open to reviewing new features and hope that ya'll find this
useful!
