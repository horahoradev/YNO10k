let saveOptions = document.getElementById("saveOptions");
let wantUploadSaveButton = document.getElementById("wantUploadSaveButton");
let uploadSaveButton = document.getElementById("uploadSaveButton");
let uploadSaveContainet = document.getElementById("uploadSaveContainer");
let uploadSaveFileInput = document.getElementById("uploadSaveFileInput");
let uploadSaveSlotInput = document.getElementById("uploadSaveSlotInput");

wantUploadSaveButton.onclick = function() {
	uploadSaveContainet.style.display = "block";
	saveOptions.style.display = "none";
}

uploadSaveButton.onclick = async function() {
	uploadSaveContainet.style.display = "none";
	saveOptions.style.display = "block";
	let saveFileBuffer = new Uint8Array(await uploadSaveFileInput.files[0].arrayBuffer());
  	FS.writeFile("Save/Save" + (uploadSaveSlotInput.value < 10 ? "0" : "") + uploadSaveSlotInput.value + ".lsd", saveFileBuffer);
	FS.syncfs(false, function(){});
}