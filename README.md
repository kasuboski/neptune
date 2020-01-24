# Neptune
Manage your places to find where to go.

## Import
Currently, it allows you to import a geojson file similar to what you get from [Google Takeout](https://takeout.google.com) for saved places.

It will also support loading from a csv file which is what you get from Takeouts for other saved places lists such as Favorites.

Neptune outputs your places to `data/out/places` where it stores a bunch of json files each with a place.

You will need a Google API Key for the Places API in order to look up more info. You can get that [here](https://developers.google.com/places/web-service/get-api-key).