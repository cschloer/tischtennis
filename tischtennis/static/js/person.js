const setup = (reporterId) => {
  $(".report-game-box button").on("click", async function () {
    const errorLine = $(".report-game-box .error-line");
    const button = $(this);
    try {
      button.prop("disabled", true);
      button.toggleClass("is-loading");
      errorLine.hide();
      errorLine.html("");

      const otherPersonIdStr = $(
        '.report-game-box select[name="other_person"]'
      ).val();
      if (!otherPersonIdStr) {
        throw new Error("Must have selected an opponent");
      }
      const otherPersonId = parseInt(otherPersonIdStr);
      const winsStr = $('.report-game-box input[name="wins"]').val();
      const wins = parseInt(winsStr);
      const lossesStr = $('.report-game-box input[name="losses"]').val();
      const losses = parseInt(lossesStr);
      const personAccessKey = $(
        '.report-game-box input[name="person_access_key"]'
      ).val();
      console.log(otherPersonId, wins, losses, personAccessKey);

      const res = await fetch(`/game`, {
        method: "POST",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
          [ACCESS_KEY_HEADER_KEY]: personAccessKey,
        },
        body: JSON.stringify({
          reporterId,
          otherPersonId,
          wins,
          losses,
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
