property childPIDs : {}

on open inputItems
	repeat with inputItem in inputItems
		launchGrafview(POSIX path of inputItem)
	end repeat
end open

on run
	display dialog "Drop a Grafana dashboard JSON or folder onto this app, or use Finder Open With." buttons {"OK"} default button "OK"
end run

on idle
	set childPIDs to livePIDs(childPIDs)
	if childPIDs is {} then quit
	return 5
end idle

on quit
	repeat with childPID in childPIDs
		do shell script "kill -TERM " & quoted form of childPID & " >/dev/null 2>&1 || true"
	end repeat
	set childPIDs to {}
	continue quit
end quit

on launchGrafview(inputPath)
	set cmd to "export PATH=/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin:/usr/sbin:/sbin:/Applications/Docker.app/Contents/Resources/bin:$HOME/go/bin; echo \"$(date): grafview " & quoted form of inputPath & " PATH=$PATH\" >> /tmp/grafview-finder.log; grafview " & quoted form of inputPath & " >> /tmp/grafview-finder.log 2>&1 & echo $!"
	set childPIDs to childPIDs & {do shell script cmd}
end launchGrafview

on livePIDs(pids)
	set liveOnes to {}
	repeat with childPID in pids
		do shell script "kill -0 " & quoted form of childPID & " >/dev/null 2>&1 && echo live || true"
		if result is "live" then set liveOnes to liveOnes & {childPID as text}
	end repeat
	return liveOnes
end livePIDs
