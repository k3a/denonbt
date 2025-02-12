#### denonbt

This is a daemon for controlling Denon 500-series receivers with Bluettoth such as `AVR-X520BT`.

I made the daemon because the original remote RC-1196 stopped working reliably and I also wanted to control the receiver from the Home Assistant.

It listens on a HTTP port and uses Bluetooth SPP (serial port) to send commands to the receiver. The commands being sent were found by using Android's HCI snoop log functionality while using their official "500 Series" Android app.

Currently only Linux is supported.
Before starting the daemon, you need to pair (bond) the Denon receiver using `bluetoothctl` or similar.
Once paired, the daemon will (re-)connect itself as necessary.
If using systemd, you may find denonbt.service useful.

Released under GNU GPL v3.
Contributions are welcome.