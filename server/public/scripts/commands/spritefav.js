
let favsprites = {};

function loadFavSprites() {
	if(window.localStorage.hasOwnProperty(gameName + 'favsprites')) {
		favsprites = JSON.parse(window.localStorage.getItem(gameName + 'favsprites'));
	}
}

function saveFavSprites() {
	window.localStorage.setItem(gameName + 'favsprites', JSON.stringify(favsprites));
}

function addFavSprite(sheet, id, name) {
	favsprites[name] = {sheet: sheet, id: parseInt(id)};
	saveFavSprites();
}

function removeFavSprite(name) {
	delete favsprites[name];
	saveFavSprites();
}

function SpriteFavCommand(args) {
	if(args.length == 4) {

		addFavSprite(args[1], args[2], args[3]);

		return true;
	}
	return false;
}

function SpriteUnfavCommand(args) {
	if(args.length == 2) {

		removeFavSprite(args[1]);

		return true;
	}

	return false;
}

loadFavSprites();