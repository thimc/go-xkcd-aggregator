# xkcd aggregator

This program fetches [xkcd.com](https://xkcd.com) comics and stores them to a
local SQLite database.

## Usage
```sh
usage: goxkcd <command> <args>
  download       - downloads and indexes all missing xkcd entries
  search <term>  - looks up any xkcd entries matching "term"
```

The output from `search` is TAB-separated which makes it very `grep`-able.

Specifying `-` as the search argument will cause program to dump all of its
entries from the database.

_Note: The `download` command will essentially DDOS xkcd.com the way it is
setup currently. The program checks the latest entry number and loops from zero
(or from when you last stopped the program) up to the latest. From what I could
read on the xkcd website there's no JSON endpoint for gathering all comics in
on request so just about 3000 HTTP requests is needed here unfortunately._


## TODO
The database layer and the business logic are currently mashed up together to
create the `xkcdstore` package which isn't ideal. the preferred solution here
would be to split them up.

The HTTP requests done to xkcd.com are using the default HTTP client using the
default configuration, having the option to specify a timeout would be nice.
