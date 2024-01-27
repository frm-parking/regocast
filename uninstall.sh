systemctl --no-relocate is-active restream.service && systemctl stop restream.service && systemctl disable restream.service && systemctl daemon-reload
rm /etc/systemd/system/restream.service

read -p "Do you want to delete config files? (Yes/No) " answer
case "$answer" in 
  [yY] | [yY][eE][sS] ) 
    rm -rf /etc/restream 
    echo "Cleaning"
    ;;
  [nN] | [nN][oO] )
    echo "Delete skipped"
    ;;
  * ) 
    echo "Invalid input"
    ;;
esac
