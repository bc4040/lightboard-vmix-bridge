# lightboard-vmix-bridge

This simple program allows light cues from an ETC EOS console to trigger graphics (or other events) on a VMIX production switcher.

The inital concept was to have light cues trigger "Scene titles" as subtitles on a VMIX switcher during a play or musical production.

## Setup your ETC EOS lighting console

Configure your ETC EOS console to send UDP Strings to the IP address of the host running this program.  The default port is 5000.

Assign a Strings value per-cue on your ETC EOS cue list.  These strings will be sent to the host when each cue is fired.

## Build with Docker

`docker build -t lightboard-vmix-bridge:latest .`

## Run with Docker
Be sure to define the IP address of VMIX as the VMIX_IP environment variable.

Bind to the host's network interface.  The listening port will be 5000/udp.

To run the program in your console:

`docker run --rm -it -e VMIX_IP=192.168.1.161 --network host lightboard-vmix-bridge`


Or, to run in the background as a service:

`docker run -d -e VMIX_IP=192.168.1.161 --network host --restart unless-stopped lightboard-vmix-bridge`

## Run the program binary directly
Alternatively, start lightboard-vmix-bridge with the IP address of VMIX as an argument:

`./lightboard-vmix-bridge 192.168.1.161`


## Predefined translations

| UDP String  | VMIX API Call |
| ------------- | ------------- |
| TOP  | Function=ScriptStart&Value=TOP |
| SCENE  | Function=ScriptStart&Value=SCENE  |
| SCN,3  | Function=DataSourceSelectRow&Value=Scenes,3<br>  Function=ScriptStart&Value=GFXSCENE  |


### Notes

There is a 6 second "cooldown" time between commands.  This allows time for an operator to go back, then forward, 1 cue without double-triggering a VMIX API call.