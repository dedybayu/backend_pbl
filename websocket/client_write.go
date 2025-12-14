package websocket

func (c *Client) WritePump() {
	defer c.Conn.Close()

	for msg := range c.Send {
		if err := c.Conn.WriteMessage(1, msg); err != nil {
			break
		}
	}
}
