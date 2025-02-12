# denonbt

This is a daemon for controlling Denon 500-series receivers with Bluettoth such as `AVR-X520BT`.

I made the daemon because the original remote RC-1196 stopped working reliably and I also wanted to control the receiver from the Home Assistant.

### Usage

It listens on a HTTP port and uses Bluetooth SPP (serial port) to send commands to the receiver. The commands being sent were found by using Android's HCI snoop log functionality while using their official "500 Series" Android app.

For the list of REST commands, please see `main.go`.

Only Linux is currently supported.
Before starting the daemon, you need to pair (bond) the Denon receiver using `bluetoothctl` or similar and then provide the receiver's MAC address as the `-hwaddr` argument.
Once paired with the system, the running daemon will (re-)connect to the receiver as necessary.
If you are using systemd, you may find denonbt.service useful.

### License and Contributing

Released under GNU GPL v3.
Contributions are welcome.
