# EWC (Easy Wifi Connection)

Perhaps we all agree that connecting or switching wifi network in Linux can be a pain. (if you're a terminal dude!)

Recently I became really frustrated with `nmcli` and its over-complicated process. So I decided to build something that I can use more easily to switch access points.

I don't know if there's already such a program that can help you to do so, I didn't research and honestly I don't care! I wanted to build it myself and also have some fun.

So here it is! A program that helps you to control you wireless connection on your Linux machine. Most important goal of this project was to be genuinely easy to use.

It looks cool (uses `huh?` library from [charmbracelet](https://github.com/charmbracelet)), it's easy to use and fast enough.

## Examples
Connecting and forgetting:

![001](https://github.com/user-attachments/assets/fbbbc235-333e-4aba-9e47-50f48b8db108)

Direct connection:

![002](https://github.com/user-attachments/assets/ac2e2aae-55ff-479a-946d-398aefdbbe2a)


Turning Wi-Fi on/off:

![002](https://github.com/user-attachments/assets/d7741139-8daf-42c9-970e-18ff198dca96)


## Features
- **Connect and Switch between Wi-Fi networks:** easily connect the available Wi-Fi networks with a simple and interactive interface
- **Save Passwords:** The program is able to save the passwords of each access point so you don't have to re-enter them every time.
- **Forget Networks:** Remove saved networks if you want to try another password.
- **Direct connection without scanning:** You can use direct connection to connect to a SSID that you already know.
- **Disable/Enable Wi-Fi:** Switch the Wi-Fi on/off with less typing!
- **User-Friendly Interface:** Built with `huh?` library, providing an intuitive and visually appealing user experience.
- **Fast and Lightweight:** It is fast enough to not waste your time; if you don't believe you can give it a shot!

## Installation
### Manual (recommended)

1. **Clone the repository:**
```
git clone https://github.com/Vistahm/ewc.git
cd ewc
```

2. **Build the program:**
 (Make sure you have [Go](https://go.dev/) installed) Run:
```
go build -o ewc *.go
```

3. **Run the program:**
 `./ewc`

 You can also move the executable file to your `/usr/local/bin` directory to use it globally on your machine.
```
mv ./ewc /usr/local/bin
```

### Auto

For auto installation just enter the following line in your terminal:
```
bash -c "$(curl -sLo- https://gist.githubusercontent.com/Vistahm/9a0d968f1e20057e534559e8e016adc6/raw/8766da4ffe7ef9c58733716319f865e69007a428/install.sh)"
```

## Dependencies

This project requires the following Go libraries:

- [huh?](https://github.com/charmbracelet/huh) - For the user interface.
- [godbus/dbus](https://github.com/godbus/dbus) - For D-Bus communication.

When you build the project, [Go](https://go.dev/) will automatically download and install these dependencies for you.

## Contributing
This program is still under development. Bugs can appear. If you encounter any problem feel free to open an issue under this repository.
Also if you have any suggestions to improve the app you can send me a meesage in Telegram (ID exists in bio) or Reddit by u/vistahm.

Any contributions are welcome!
