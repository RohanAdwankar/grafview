on open inputItems
	repeat with inputItem in inputItems
		set inputPath to POSIX path of inputItem
		set cmd to "export PATH=/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin:/usr/sbin:/sbin:/Applications/Docker.app/Contents/Resources/bin:$HOME/go/bin; echo \"$(date): grafview " & quoted form of inputPath & " PATH=$PATH\" >> /tmp/grafview-finder.log; grafview " & quoted form of inputPath & " >> /tmp/grafview-finder.log 2>&1 &"
		do shell script cmd
	end repeat
end open

on run
	display dialog "Drop a Grafana dashboard JSON or folder onto this app, or use Finder Open With." buttons {"OK"} default button "OK"
end run
