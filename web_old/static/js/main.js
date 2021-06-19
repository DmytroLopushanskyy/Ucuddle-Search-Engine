// Search box

const words = ["learn python", "is python relevant today", "python online courses", "Погода Тернопіль", "Яка буде погода у Івано-Франківську", "Чи погіршиться погода?", "Погода у Львові", "suppose end get","boy warrant general","natural. delightful","met sufficient projection ask.","decisively everything","principles if preference","do","impression","of. preserved oh so","difficult repulsive on","in household. in what","do","miss time be. valley","as be","appear","cannot so","by.","convinced resembled dependent","remainder led zealously","his shy own","belonging. always length","letter","adieus","add number moment she.","promise few","compass six several old","offices removal parties","fat. concluded","rapturous it intention","perfectly daughters","is as.","behaviour we","improving at something","to. evil true","high lady roof men","had open.","to projection considered it","precaution an","melancholy or.","wound young","you thing","worse along being ham.","dissimilar of favourable solicitude","if sympathize middletons","at. forfeited","up if disposing","perfectly in an","eagerness perceived necessary.","belonging sir","curiosity discovery","extremity yet","forfeited prevailed","own off.","travelling by","introduced of","mr terminated. knew as","miss","my high hope quit. in","curiosity shameless dependent","knowledge up.","literature admiration","frequently indulgence announcing","are who you","her. was","least quick after","six. so","it yourself repeated","together","cheerful. neither it cordial so","painful picture studied if.","sex","him position doubtful","resolved boy expenses.","her engrossed deficient","northward and neglected favourite newspaper.","but","use peculiar","produced concerns ten. maids","table how","learn","drift","but purse","stand yet","set. music me","house could","among oh","as their. piqued","our","sister shy nature almost","his","wicket.","hand dear","so we hour","to. he","we be","hastily","offence effects","he service. sympathize","it projection","ye insipidity celebrated"]

const containerEl = document.querySelector('.container')
const formEl = document.querySelector('#search')
const dropEl = document.querySelector('.drop')

const formHandler = (e) => {
    const userInput = e.target.value.toLowerCase()

    if(userInput.length === 0) {
        dropEl.style.height = 0
        return dropEl.innerHTML = ''              
    }

    const filteredWords = words.filter(word => word.toLowerCase().includes(userInput)).sort().splice(0, 5)
    
    dropEl.innerHTML = ''
    filteredWords.forEach(item => {
        const listEl = document.createElement('li')
        var link = document.createElement('a');
        var linkText = document.createTextNode(item);
        link.appendChild(linkText);
        link.href = "/search?query=" + item;
        listEl.appendChild(link);
        if(item === userInput) {
            listEl.classList.add('match');
        }
        dropEl.appendChild(listEl);
    })

    if(dropEl.children[0] === undefined) {
        return dropEl.style.height = 0
    }

    let totalChildrenHeight = dropEl.children[0].offsetHeight * filteredWords.length
    dropEl.style.height = totalChildrenHeight + 'px'

}

formEl.addEventListener('input', formHandler)

// Images

function getRandomImage() {
  var images = ["1.webp","2.webp","3.webp", "4.webp","5.webp","6.webp"];

  return images[Math.floor(Math.random() * images.length)];
}



$(document).ready(function(){
    $("body").fadeIn(500);
    $("#main-page").css("backgroundImage",'linear-gradient( rgba(234, 124, 34, 0.1), rgba(234, 124, 34, 0.1) ), url("../static/img/' + getRandomImage() + '")');
});


// Load links

$(document).on('click', '#more', function(e) {
    e.preventDefault()

    var start = $('#more').attr('start')
    var end = $('#more').attr('end')
    var search = $('#more').attr('search')

    $.ajax({
        url: '/more_links',
        data: JSON.stringify({'search': search, 'start':start, 'end': end}, null, '\t'),
        type: 'POST',
        contentType: 'application/json;charset=UTF-8',
        success: function(response) {
            console.log(response);
            for (var i = 0; i < response.websites.length; i++) {
                $('.content').append("<li><span>" + response.websites[i].link.slice(0, 40) + "</span><a href='" + response.websites[i].link + "'><h2>" + response.websites[i].title + "</h2></a><p>" + response.websites[i].description + "</p></li>")
            }
        },
        error: function(error) {
            console.log(error);
        }
    });

});


// Form 

$(function() {
    $('.search-container').each(function() {
        $(this).find('input').keypress(function(e) {
            // Enter pressed?
            console.log("hello");
            if(e.which == 10 || e.which == 13) {
                window.location.href = "/search?query=" + $("#search").val();
            }
        });
    });
});




