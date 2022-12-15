const ACCESS_KEY_HEADER_KEY = "X-Wall-City-Access-Key";

const checkRes = async (res) => {
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text);
  }
  return true;
};
