const setup = () => {
  $(".create-person-box button").on("click", async function () {
    const errorLine = $(".create-person-box .error-line");
    const button = $(this);
    try {
      button.prop("disabled", true);
      button.toggleClass("is-loading");
      errorLine.hide();
      errorLine.html("");

      const name = $('.create-person-box input[name="name"]').val();
      if (!name) {
        throw new Error("Person must have a name");
      }
      const faIcon = $('.create-person-box input[name="fa_icon"]').val();
      console.log("FA ICON");

      const personAccessKey = $(
        '.create-person-box input[name="person_access_key"]'
      ).val();
      if (!personAccessKey) {
        throw new Error("Person must have an access key");
      }
      const adminAccessKey = $(
        '.create-person-box input[name="admin_access_key"]'
      ).val();

      const res = await fetch(`${BASE_PATH}/person`, {
        method: "POST",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
          [ACCESS_KEY_HEADER_KEY]: adminAccessKey,
        },
        body: JSON.stringify({
          name,
          personAccessKey,
          faIcon,
        }),
      });
      await checkRes(res);

      window.onbeforeunload = () => {};
      window.location.reload(false);
    } catch (error) {
      errorLine.show();
      errorLine.html(error);
    } finally {
      button.toggleClass("is-loading");
      button.prop("disabled", false);
    }
  });
};

const deletePerson = async (personId) => {
  const adminAccessKey = prompt("Please enter the admin access key");
  if (adminAccessKey) {
    const button = $(`.person-delete-${personId}`);
    const loader = $(`.person-loader-${personId}`);
    const errorLine = $(".person-error-line");
    try {
      button.hide();
      loader.show();
      errorLine.hide();
      errorLine.html("");
      const res = await fetch(`${BASE_PATH}/person/`, {
        method: "DELETE",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
          [ACCESS_KEY_HEADER_KEY]: adminAccessKey,
        },
        body: JSON.stringify({
          personId,
        }),
      });
      await checkRes(res);

      window.onbeforeunload = () => {};
      window.location.reload(false);
    } catch (error) {
      errorLine.show();
      errorLine.html(error);
    } finally {
      button.show();
      loader.hide();
    }
  }
};

const deleteGame = async (personId, created) => {
  const adminAccessKey = prompt("Please enter the admin access key");
  if (adminAccessKey) {
    const button = $(`.game-delete-${personId}-${created}`);
    const loader = $(`.game-loader-${personId}-${created}`);
    const errorLine = $(".game-error-line");
    try {
      button.hide();
      loader.show();
      errorLine.hide();
      errorLine.html("");
      const res = await fetch(`${BASE_PATH}/game`, {
        method: "DELETE",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
          [ACCESS_KEY_HEADER_KEY]: adminAccessKey,
        },
        body: JSON.stringify({
          personId,
          created,
        }),
      });
      await checkRes(res);

      window.onbeforeunload = () => {};
      window.location.reload(false);
    } catch (error) {
      errorLine.show();
      errorLine.html(error);
    } finally {
      button.show();
      loader.hide();
    }
  }
  return false;
};
