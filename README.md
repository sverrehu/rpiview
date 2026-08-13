# rpiview

Intention: Show images from a Raspberry Pi camera on the local computer.

Works by opening an image window and starting a web server to receive images.
Given an image on the Raspberry Pi, you may POST it as binary data to the web server, which will display it in the image window.
The URL to POST images to will be displayed upon start-up.

Please note: This is only to be run in a local network.
Do not expose the web server to the Internet.
