const ACCESS_KEY_HEADER_KEY = "X-Wall-City-Access-Key";

const checkRes = async (res) => {
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text);
  }
  return true;
};

const copyToClipboard = (val) => {
  console.log("INSIDE COPY TO CLIPBOARD");
  var dummy = document.createElement("input");
  dummy.style.display = "none";
  document.body.appendChild(dummy);

  dummy.setAttribute("id", "dummy_id");
  document.getElementById("dummy_id").value = val;
  dummy.select();
  document.execCommand("copy");
  document.body.removeChild(dummy);
};
