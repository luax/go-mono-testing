var socket = new WebSocket(
  (window.location.protocol === "https:" ? "wss://" : "ws://") +
    window.location.host +
    "/ws"
);
socket.onopen = () => {
  console.log("open");
  socket.send(
    JSON.stringify({
      data: "foo bar " + Math.random()
    })
  );
};
socket.onmessage = e => {
  console.log("Message: ", e.data);
};
socket.onclose = () => {
  console.log("close");
};
