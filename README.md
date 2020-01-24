# Neptune
Manage your places to find where to go.

## Install
`go get -u github.com/kasuboski/neptune`

You will need a Google API Key for the Places API in order to look up more info. You can get that [here](https://developers.google.com/places/web-service/get-api-key).

## Import
Export your Google Saved Places from [Google Takeout](https://takeout.google.com).

This will get you a `json` file and multiple `csv` files. The `json` file is your starred places and the `csv` is for other lists e.g. Want to go, Favorites.

Import the `json` file with `neptune import geojson -f /path/to/file.json -t saved`
Import each of the `csv` files with `neptune import csv -f /path/to/file.csv -t <listname>`

Neptune outputs your places to `data/out/places` where it stores a bunch of json files each with a place.