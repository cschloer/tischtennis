const ACCESS_KEY_HEADER_KEY = "X-Wall-City-Access-Key";

const checkRes = async (res) => {
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text);
  }
  return true;
};

function copyToClipboard(text) {
  var dummy = document.createElement("textarea");
  // to avoid breaking orgain page when copying more words
  // cant copy when adding below this code
  // dummy.style.display = 'none'
  document.body.appendChild(dummy);
  //Be careful if you use texarea. setAttribute('value', value), which works with "input" does not work with "textarea". – Eduard
  dummy.value = text;
  dummy.select();
  document.execCommand("copy");
  document.body.removeChild(dummy);
  return false;
}
