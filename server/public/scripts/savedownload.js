let wantDownloadSaveButton = document.getElementById("wantDownloadSaveButton");
let downloadSaveButton = document.getElementById("downloadSaveButton");
let dowloadSaveContainer = document.getElementById("downloadSaveContainer");
let dowloadSaveSlotInput = document.getElementById("downloadSaveSlotInput");

wantDownloadSaveButton.onclick = function() {
	downloadSaveContainer.style.display = "block";
	saveOptions.style.display = "none";
}

downloadSaveButton.onclick = async function() {

	let slotname = "Save" + (downloadSaveSlotInput.value < 10 ? "0" : "") + downloadSaveSlotInput.value + ".lsd";

	let blob = new Blob([FS.readFile("Save/" + slotname)]);
	let link = document.createElement("a");
	link.href = window.URL.createObjectURL(blob);
	link.download = slotname;
	link.click();
	saveOptions.style.display = "block";
	downloadSaveContainer.style.display = "none";
}