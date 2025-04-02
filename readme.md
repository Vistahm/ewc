# EWC (Easy Wifi Connection)

Perhaps we all agree that connecting or switching wifi network in Linux can be a pain. (if you're a terminal dude!)

Recently I became really frustrated with all this `nmcli` thing and I decided to build something that I can use more easily to switch access points and turn the wifi on or off.

I don't know if there's already such a program that can help you to do so, I didn't research and honestly I don't care! I wanted to build it myself and also have some fun.

So here it is! A program that helps you to control you wireless connection on your Linux machine.

It looks cool (uses `huh?` library from charmbracelet), it's easy to use and it is fast enough.

## Features
- **Connect and Switch between Wi-Fi networks:** easily connect the available Wi-Fi networks with a simple and interactive interface
- **Save Passwords:** The program is able to save the passwords of each access point so you don't have to re-enter them every time.
- **Forget Networks:** Remove saved networks if you want to try another password.
- **User-Friendly Interface:** Built with `huh?` library, providing an intuitive and visually appealing user experience.
- **Fast and Lightweight:** It is fast enough to not waste your time; if you don't believe you can give it a shot!

## Installation
1. **Clone the repository:**
```
git clone https://github.com/Vistahm/ewc.git
cd ewc
```

2. **Build the program:**
 (Make sure you have Go installed) Run:
 `go build -o ewc *.go`

3. **Run the program:**
 `./ewc`

 You can also move the executable file to your `/usr/local/bin` directory to use it globally on you machine.
 `mv ./ewc /usr/local/bin`

## Dependencies

This project requires the following Go libraries:

- [huh?](https://github.com/charmbracelet/huh) - For the user interface.
- [godbus/dbus](https://github.com/godbus/dbus) - For D-Bus communication.

When you build the project, Go will automatically download and install these dependencies for you.

## Contributing
This program is still under some development. There can be found some potentially bugs. If you encounter any problem feel free to open an issue under this repository.

Any Contributions are welcome!
