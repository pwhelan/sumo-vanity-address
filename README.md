# sumo-vanity-address

A Sumo vanity address generator in Go(lang).

This nitfy little program attempts to generate SUMO addresses that begin with a specific set of characters.
To compile it you need the golang compiler and [gb](https://getgb.io/). Once those are installed:

    $ git clone https://github.com/pwhelan/sumo-vanity-address/
    $ cd sumo-vanity-address
    $ gb vendor fetch
    $ gb build

To invoke:

    $ ./bin/vanity LoL

To also choose the leading number:

    $ ./bin/vanity --numeral=3 LoL

Screenshot:

![screenshot](https://raw.githubusercontent.com/pwhelan/sumo-vanity-address/master/vanity.png)

Donation address: Sumoo4HoDLxbrFBuQyXubwjfjQLUteV4K2PQ6VE62UESBVfr5rv9yzHbTVX246aZAQ1nMvqmySyq73Tdsv12N4vbChacH3sDHrT
