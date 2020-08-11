for ((i=1;i<=100;i++));
do
   # your-unix-command-here
   echo $i
   filename=$( printf 'text/%d.txt' $i )
   echo 'This is a test $filename' > $filename
   git add . -A
   git commit -m "$filename"
   git push
done
