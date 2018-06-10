var socket = new WebSocket("/ws");
socket.onopen = () => {
  console.log("open");
  socket.send(
    JSON.stringify({
      data: "foo bar"
    })
  );
};
socket.onmessage = e => {
  console.log(e);
};
socket.onclose = () => {
  console.log("close");
};
